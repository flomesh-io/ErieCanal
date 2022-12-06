pipy()
  .listen(os.env.PIPY_SIDECAR_HTTP_PORT)
  .connect(() => os.env.PIPY_SERVICE_HTTP_ADDR || "127.0.0.1:8080")
  .listen(os.env.PIPY_HELLO_HTTP_PORT)
  .decodeHttpRequest()
  .replaceMessage(
    new Message('Hello!\n')
  )
  .encodeHttpResponse()