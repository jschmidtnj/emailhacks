'use strict'

// credit to https://medium.com/@felice.geracitano/brotli-compression-delivered-from-aws-7be5b467c2e1
// https://medium.com/@kazaz.or/aws-cloudfront-compression-using-lambda-edge-where-is-brotli-6d296f41f784
// https://github.com/CloudUnder/lambda-edge-nice-urls

const regexSuffixless = /\/[^/.]+$/; // e.g. "/some/page" but not "/", "/some/" or "/some.jpg"
const regexTrailingSlash = /.+\/$/; // e.g. "/some/" or "/some/page/" but not root "/"

exports.handler = (event, context, callback) => {
  const request = event.Records[0].cf.request
  const headers = request.headers
  const useBrotli = headers['accept-encoding'] && headers['accept-encoding'][0].value.indexOf('br') > -1
  request.uri = (useBrotli ? '/brotli' : '/gzip') + request.uri
  // Append ".html" to origin request
	if (request.uri.match(regexSuffixless)) {
		request.uri = request.uri + '.html'
		callback(null, request)
		return
	}
	// Append "index.html" to origin request
	if (request.uri.match(regexTrailingSlash)) {
		request.uri = request.uri + 'index.html'
		callback(null, request)
		return
	}
	// Redirect (301) non-root requests ending in "/" to URI without trailing slash
	if (request.uri.match(/.+\/$/)) {
		const response = {
			// body: '',
			// bodyEncoding: 'text',
			headers: {
				'location': [{
					key: 'Location',
					value: request.uri.slice(0, -1)
				 }]
			},
			status: '301',
			statusDescription: 'Moved Permanently'
		}
		callback(null, response)
		return
	}
  callback(null, request)
}
