# golanding

This project is purposed to give some base for landing page serving.

How yo use it you can check it in own blog - [Serve landing page with Go Iris backend on CentOS](http://borchevkin.com/serve-landing-page-with-go/)

## License

BSD-2-Clause. Please see LICENSE file

## Who develop this

[Danil Borchevkin](https://github.com/DanilBorchevkin)

## What included

* ***main.go*** - main itself

* ***.env.example*** - example .env file for exporting enviroment variables (please see below)

* ***/comingsoon/*** - coming soon page for serving. It may use only as example of working

## Bootstrap

Please check out the article [Serve landing page with Go Iris backend on CentOS](http://borchevkin.com/serve-landing-page-with-go/)

### Setup .env file

For using this code you should copy ***.env.example*** as ***.env*** and fill following variables:

* ***STATIC_PATH*** - path where you store your landing page. If you have no suitable landing page you can use proposed *comingsoon* page with path *./comingsoon/*

* ***HTTP_PORT*** - port on which will expose a landing page. Usually 80

* ***UPLOAD_PATH*** - path for uploads files. Don't forget create this folder!

* ***SENDGRID_API_KEY*** - it's very obiviously

* ***EMAIL_ADDRESS*** - address which wil be use as *FROM* and *TO* in emails

* ***DEBUG_LEVEL*** - debug level of the Iris framework. Equals to *debug* as default

### Start program

As a peace a cake:

```shell
    go get
    go run main.go
```
.
## Routes and data format

The app has only two routes:

* ***/*** - at this route serves static content which placed at ***STATIC_PATH*** variable

* ***/createlead*** - route for post data from form (values and file). Please check *comingsoon/index.html* for get some information about it.

## Warnings and useful advices

* Files in ***UPLOAD_PATH*** should be erased manually or by cron events

## Feedback

* tech questions: [Danil Borchevkin](https://github.com/DanilBorchevkin)

* administrative questions: [Danil Borchevkin](https://github.com/DanilBorchevkin)
