'use strict'

const fs = require('fs')
const path = require('path')
const AWS = require('aws-sdk')
const zlib = require('zlib')
const mime = require('mime-types')
require('dotenv').config()

AWS.config = new AWS.Config()
AWS.config.accessKeyId = process.env.AWS_ACCESS_KEY_ID
AWS.config.secretAccessKey = process.env.AWS_SECRET_ACCESS_KEY
AWS.config.region = process.env.AWS_REGION

const sourcePath = process.env.SOURCE_DIR
const bucketName = process.env.AWS_S3_BUCKET
const cloudfrontID = process.env.AWS_CLOUDFRONT_ID

const gzipBase = 'gzip'
const brotliBase = 'brotli'

const s3Client = new AWS.S3({
  params: {
    Bucket: bucketName
  }
})

const cloudfrontClient = new AWS.CloudFront({
  params: {
    DistributionId: cloudfrontID
  }
})

const deleteObjects = (callback) => {
  s3Client.listObjects({}, (err, data) => {
    if (err)
      return callback(err)
    for (let i = 0; i < data.Contents.length; i++) {
      s3Client.deleteObject({
        Key: data.Contents[i].Key
      }, (err) => {
        if (err)
          return callback(err)
      })
    }
    if (data.IsTruncated) {
      deleteObjects((err) => {
        if (err)
          return callback(err)
      })
    } else {
      return callback(null)
    }
  })
}
const processFile = (filePath, callback) => {
  let fileType = mime.lookup(path.extname(filePath))
  if (!fileType) {
    fileType = 'application/octet-stream'
  }
  const bucketPath = filePath.split(sourcePath)[1]
  const readGzip = fs.createReadStream(filePath)
  const gzipFile = readGzip.pipe(zlib.createGzip())
  let numUploads = 0
  s3Client.upload({
    Body: gzipFile,
    Key: gzipBase + bucketPath,
    ContentEncoding: 'gzip',
    ContentType: fileType
  })
    .on('httpUploadProgress', (evt) => {
      console.log(evt)
    })
    .send((err) => {
      if (err)
        return callback(err)
      numUploads++
      if (numUploads === 2) {
        callback(null)
      }
    })
  const readBrotli = fs.createReadStream(filePath)
  const brotliFile = readBrotli.pipe(zlib.createBrotliCompress())
  s3Client.upload({
    Body: brotliFile,
    Key: brotliBase + bucketPath,
    ContentEncoding: 'br',
    ContentType: fileType
  })
    .on('httpUploadProgress', (evt) => {
      console.log(evt)
    })
    .send((err) => {
      if (err)
        return callback(err)
      numUploads++
      if (numUploads === 2) {
        callback(null)
      }
    })
}
const processDirectory = (dir, callback) => {
  fs.readdir(dir, (err, list) => {
    if (err)
      return callback(err)
    let numRemaining = list.length
    if (!numRemaining)
      return callback(null)
    list.forEach((file) => {
      file = path.resolve(dir, file)
      fs.stat(file, (err, stat) => {
        if (err)
          return callback(err)
        if (stat && stat.isDirectory()) {
          processDirectory(file, (err) => {
            if (err)
              return callback(err)
            if (!--numRemaining)
              callback(null)
          })
        } else {
          processFile(file, (err) => {
            if (err)
              return callback(err)
            if (!--numRemaining)
              callback(null)
          })
        }
      })
    })
  })
}
const clearCloudfront = (callback) => {
  cloudfrontClient.createInvalidation({
    InvalidationBatch: {
      CallerReference: new Date().getTime().toString(),
      Paths: {
        Quantity: 1,
        Items: [
          '/*'
        ]
      }
    }
  }, (err) => {
    if (err)
      return callback(err)
    return callback(null)
  })
}
deleteObjects((err) => {
  if (err)
    throw err
  processDirectory(sourcePath, (err) => {
    if (err)
      throw err
    clearCloudfront((err) => {
      if (err)
        throw err
    })
  })
})
