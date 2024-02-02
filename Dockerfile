FROM golang:latest
 
 WORKDIR /app
  
  COPY go.mod .
  COPY *.go .
   
   RUN go get
   RUN CGO_ENABLED=0 GOOS=linux go build -o /go-dice-roller
    
    EXPOSE 8080
     
     ENTRYPOINT ["/go-dice-roller"]
