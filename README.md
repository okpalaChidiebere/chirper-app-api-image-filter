## chirper-app-api-image-filter

This is a microservice responsible for everything images for the chirper-app. Its runs a uses the [connect-go](https://connect.build/docs/go/getting-started/) library that runs the server that serves gRPC and HTTP/1.1 and HTTP/2.

## Starting server

- Clone this repo and run `go run main.go`.
- You can access the backend endpoint at [http://localhost:9000](http://localhost:9000); This connection is insecure though.
- It is recommended to access this backend through `chirper-app-reverse-proxy`. We do this easily with the help of [`docker compose`](https://docs.docker.com/compose/compose-file/compose-file-v3/). When you access the backend through `chirper-app-reverse-proxy` you must go through port `1443` to connect to the users microservices securely. Eg: [https://localhost:1443/user.v1.UserService/ListUsers](https://localhost:1443/user.v1.UserService/ListUsers). Port 1443 on supports HTTP/2.
- If you connect to the this microservice through port **9000** the connection will be insecure as there is no SSL and will support only HTTP/1.1; but if you go through port **1443** it uses SSL and supports only HTTP/2.
- For grPC Reflection, you will need to load the refection in postman from the insecure port `localhost:9000` in the 'new > gRPC Request' tab. After you have load the reflection, it does not matter which port us use to test all the services exposed by the reflection. The only gotcha is if you are to you want to use the secure port, you will need to upload your server cert and key and well as your Authority cert to postman from the preference screen of the app. Learn more about reflection [here](https://www.youtube.com/watch?v=yluYiCj71ss). See this [blog](https://learning.postman.com/docs/sending-requests/certificates/) on how to add SSL to postman; For me i uploaded authority cert generated from [Openssl](https://man.openbsd.org/openssl.1#x509) for the 'CA Certificates' section, server cert and server key for the 'Client Certificates' section.

## Server services

- The three main services for the demo of this project for tweets is defined [here](https://github.com/okpalaChidiebere/chirper-app-apis/blob/master/user/v1/api.proto)
- Read this [documentation](https://cloud.google.com/endpoints/docs/grpc/transcoding) to see furthermore on how to interpret the api definitions
- **NOTE** for the project, the image-filter service gives you a timeout on when you can download an image, but if you want to just open up the url, you will need to do that in the configs of the S3 bucket. This [link](https://repost.aws/knowledge-center/s3-static-website-endpoint-error) will help you understand

## Useful links about Connect-go

- [Using connect-go client in a react app](https://crieit.net/posts/connect-go-with-cors)
- [Getting started with Connect-go](https://connect.build/docs/go/getting-started/)
- [Streaming](https://connect.build/docs/go/streaming/). I did not test streaming for web clients but i am quite sure the grpc backend works event though i did not implement that

## Useful information about the CI build

- We used Travis CI for our build which basically spins up a computer for use remotely and build our app. That computer has git in it. So just provided our github credentials to it which our the computer to build our app with the private modules. It was a good learning. Now if you want do that github step in docker you can checkout this [link](https://jwenz723.medium.com/fetching-private-go-modules-during-docker-build-5b76aa690280). Remember there are ways to provide credentials to github for it to use. See them [here](https://docs.travis-ci.com/user/private-dependencies/). I prefer to use API token
- Learn how to make a go private module with docker [here](https://medium.com/the-godev-corner/how-to-create-a-go-private-module-with-docker-b705e4d195c4)
