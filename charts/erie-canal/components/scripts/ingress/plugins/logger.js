(config =>

  pipy({
    _enabled: (os.env.ENABLE_LOG && os.env.ENABLE_LOG.toLowerCase() === 'true' || config.enabled),
    _logURL: (os.env.ENABLE_LOG || config.enabled) && new URL(os.env.LOGURL || config.logURL),
    _authorization: (os.env.LOG_AUTHROIZATION || config.Authorization),
    _request: null,
    _requestTime: 0,
    _responseTime: 0,
    _responseContentType: '',
    _instanceName: os.env.HOSTNAME,

    _CONTENT_TYPES: {
      '': true,
      'text/plain': true,
      'application/json': true,
      'application/xml': true,
      'multipart/form-data': true,
    },

  })

  .import({
      __apiID: 'router',
      __ingressID: 'router',
      __projectID: 'router'
    })


  .pipeline('request')
    .link('req', () => _enabled === true, 'bypass')
    .fork('log-request')

  .pipeline('response')
    .link('resp', () => _enabled === true, 'bypass')

  .pipeline('req')
    .fork('log-request')

  .pipeline('resp')
    .fork('log-response')

  .pipeline('log-request')
    .handleMessageStart(
      () => _requestTime = Date.now()
    )
    .decompressHTTP()
    .handleMessage(
      '1024k',
      msg => (
        _request = msg
      )
    )

  .pipeline('log-response')
    .handleMessageStart(
      msg => (
        _responseTime = Date.now(),
        _responseContentType = (msg.head.headers && msg.head.headers['content-type']) || ''
      )
    )
    .decompressHTTP()
    .replaceMessage(
      '1024k',
      msg => (
        new Message(
          JSON.encode({
            req: {
              ..._request.head,
              body: _request.body.toString(),
            },
            res: {
              ...msg.head,
              body: Boolean(_CONTENT_TYPES[_responseContentType]) ? msg.body.toString() : '',
            },
            x_parameters: { aid: __apiID, igid: __ingressID, pid: __projectID },
            instanceName: _instanceName,
            reqTime: _requestTime,
            resTime: _responseTime,
            endTime: Date.now(),
            reqSize: _request.body.size,
            resSize: msg.body.size,
            remoteAddr: __inbound?.remoteAddress,
            remotePort: __inbound?.remotePort,
            localAddr: __inbound?.localAddress,
            localPort: __inbound?.localPort,
          }).push('\n')
        )
      )
    )
    .merge('log-send', () => '')

  .pipeline('log-send')
    .pack(
      1000,
      {
        timeout: 5,
        interval: 5,
      }
    )
    .replaceMessageStart(
      () => new MessageStart({
        method: 'POST',
        path: _logURL.path,
        headers: {
          'Host': _logURL.host,
          'Authorization': _authorization,
          'Content-Type': 'application/json',
        }
      })
    )
    .encodeHTTPRequest()

    .connect(
      () => _logURL.host,
      {
        bufferLimit: '8m',
      }
    )

  .pipeline('bypass')

)(JSON.decode(pipy.load('config/logger.json')))
