package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Lead struct {
	ProjectType  string
	WorkType     string
	WhenStart    string
	Fio          string
	PhoneOrSkype string
	Email        string
	Description  string
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	port := os.Getenv("PORT")
	//sendGridAPIKey := os.Getenv("SENDGRID_API_KEY")
	//emailAddress := os.Getenv("EMAIL_ADDRESS")
	staticPath := os.Getenv("STATIC_PATH")
	debugLevel := os.Getenv("DEBUG_LEVEL")

	app := iris.New()
	app.Logger().SetLevel(debugLevel)
	app.Use(recover.New())
	app.Use(logger.New())
	app.StaticWeb("/", staticPath)
	app.OnErrorCode(iris.StatusNotFound, notFoundHandler)
	app.OnErrorCode(iris.StatusInternalServerError, internalServerErrorHandler)
	app.Post("/createlead", iris.LimitRequestBodySize(10<<20), createLeadHandler)

	app.Run(iris.Addr(":"+port), iris.WithoutServerError(iris.ErrServerClosed))
}

func notFoundHandler(ctx iris.Context) {
	// when 404 then render the template $views_dir/errors/404.html
	ctx.WriteString("Ooops. 404 error")
}

func internalServerErrorHandler(ctx iris.Context) {
	ctx.WriteString("Oups something went wrong, try again")
}

func createLeadHandler(ctx iris.Context) {
	lead := Lead{}

	err := ctx.ReadForm(&lead)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{
			"status":  400,
			"message": err.Error(),
		})
		return
	}

	// Get the file from the request
	file, info, err := ctx.FormFile("File")
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{
			"status":  400,
			"message": err.Error(),
		})
		return
	}
	defer file.Close()

	timestamp := strconv.FormatInt((time.Now().UnixNano() / 1e6), 10)
	fpath := os.Getenv("UPLOAD_PATH") + timestamp + "_" + info.Filename

	// Create a file with the same name
	// assuming that you have a folder named 'uploads'
	out, err := os.OpenFile(fpath,
		os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{
			"status":  400,
			"message": err.Error(),
		})
		return
	}
	defer out.Close()

	io.Copy(out, file)

	err = sendLeadByEmail(lead, fpath)
	if err != nil {
		ctx.StatusCode(iris.StatusBadRequest)
		ctx.JSON(iris.Map{
			"status":  400,
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(iris.Map{
		"status":  200,
		"message": "A lead successfully created",
		"lead":    lead,
	})
}

func sendLeadByEmail(lead Lead, fpath string) error {
	request := sendgrid.GetRequest(os.Getenv("SENDGRID_API_KEY"), "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"

	body, err := getSendgridBody(lead, fpath)
	if err != nil {
		return err
	}

	request.Body = body

	response, err := sendgrid.API(request)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}

	return err
}

func getSendgridBody(lead Lead, fpath string) ([]byte, error) {
	emailAddress := os.Getenv("EMAIL_ADDRESS")

	const tpl = `
		<p>Project type: {{.ProjectType}}</p>
		<p>Work type: {{.WorkType}}</p>
		<p>When start: {{.WhenStart}}</p>
		<p>Fio: {{.Fio}}</p>
		<p>Phone or skype: {{.PhoneOrSkype}}</p>
		<p>Email: {{.Email}}</p>
		<p>Description: {{.Description}}</p>
	`

	t, err := template.New("webpage").Parse(tpl)
	if err != nil {
		return nil, err
	}

	var buff bytes.Buffer
	err = t.Execute(&buff, lead)
	if err != nil {
		return nil, err
	}
	htmlData := buff.String()

	f, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}

	_, filename := filepath.Split(fpath)

	m := mail.NewV3Mail()
	e := mail.NewEmail(emailAddress, emailAddress)
	m.SetFrom(e)

	p := mail.NewPersonalization()
	tos := []*mail.Email{
		mail.NewEmail(emailAddress, emailAddress),
	}
	p.AddTos(tos...)

	m.AddPersonalizations(p)

	m.Subject = "New Lead - " + lead.Fio + " / " + lead.Email + " / " + lead.PhoneOrSkype

	c := mail.NewContent("text/plain", "You got new lead")
	m.AddContent(c)

	c = mail.NewContent("text/html", htmlData)
	m.AddContent(c)

	attachedFile := mail.NewAttachment()
	attachedFile.SetContent(base64.StdEncoding.EncodeToString(f))
	attachedFile.SetFilename(filename)
	attachedFile.SetDisposition("attachment")
	m.AddAttachment(attachedFile)

	return mail.GetRequestBody(m), err
}

