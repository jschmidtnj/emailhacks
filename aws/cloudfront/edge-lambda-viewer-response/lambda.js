'use strict'
// credit to https://medium.com/@tom.cook/edge-lambda-cloudfront-custom-headers-3d134a2c18a2
exports.handler = (event, context, callback) => {
  const response = event.Records[0].cf.response
  const headers = response.headers
  headers['strict-transport-security'] = [{
    key: 'Strict-Transport-Security',
    value: 'max-age=31536000; includeSubdomains; preload'
  }]
  /*
  headers['content-security-policy'] = [{
    key: 'Content-Security-Policy',
    value: 'default-src 'none'; img-src 'self'; script-src 'self'; style-src 'self'; object-src 'none''
  }]
  */
  headers['x-content-type-options'] = [{
    key: 'X-Content-Type-Options',
    value: 'nosniff'
  }]
  headers['x-frame-options'] = [{
    key: 'X-Frame-Options',
    value: 'DENY'
  }]
  headers['x-xss-protection'] = [{
    key: 'X-XSS-Protection',
    value: '1; mode=block'
  }]
  headers['referrer-policy'] = [{
    key: 'Referrer-Policy',
    value: 'same-origin'
  }]
  const ending = response.uri.match(/\.[0-9a-z]+$/i)
  let cacheTime
  switch (ending) {
    case '.js':
      cacheTime = 31536000
      break
    case '.css':
      cacheTime = 31536000
      break
    default:
      cacheTime = 21536000
      break
  }
  headers['cache-control'] = [{
    key: 'cache-control',
    value: `public, max-age=${cacheTime}`
  }]
  headers['vary'] = [{
    key: 'vary',
    value: 'accept-encoding'
  }]
  callback(null, response)
}
