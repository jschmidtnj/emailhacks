'use strict'
// credit to https://medium.com/@felice.geracitano/brotli-compression-delivered-from-aws-7be5b467c2e1
// https://medium.com/@kazaz.or/aws-cloudfront-compression-using-lambda-edge-where-is-brotli-6d296f41f784
exports.handler = (event, context, callback) => {
  const request = event.Records[0].cf.request
  const headers = request.headers
  const useBrotli = headers['accept-encoding'] && headers['accept-encoding'][0].value.indexOf('br') > -1
  request.uri = (useBrotli ? '/brotli' : '/gzip') + request.uri
  callback(null, request)
}
