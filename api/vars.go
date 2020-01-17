package main

import (
	"time"
)

var mainDatabase = "website"

var userMongoName = "users"

var responseMongoName = "responses"

var formMongoName = "forms"

var projectMongoName = "projects"

var blogMongoName = "blogs"

var shortLinkMongoName = "shortlink"

type key string

const tokenKey key = "token"

const getTokenKey key = "misc"

const dataKey key = "data"

const getConnectionIDKey key = "connection"

var graphiQL = false

var graphqlPlayground = true

var sitemapTimeFormat = "2006-01-02T15:04:05Z07:00"

var dateFormat = "Mon Jan _2 15:04:05 2006"

var hexRegex = "(^#[0-9A-F]{6}$)|(^#[0-9A-F]{8}$)"

var formType = "form"

var projectType = "project"

var responseType = "response"

var blogType = "blog"

var validStorageTypes = []string{
	formType,
	responseType,
	blogType,
}

var validOrganization = []string{
	"category",
	"tag",
}

var superAdminType = "super"

var adminType = "admin"

var userType = "user"

var validAccessTypes = []string{
	"admin",
	"edit",
	"view",
	"shared",
	"none",
}

var sharedAccessLevel = validAccessTypes[3]
var noAccessLevel = validAccessTypes[4]

var editAccessLevel = []string{
	validAccessTypes[0],
	validAccessTypes[1],
}

var viewAccessLevel = []string{
	validAccessTypes[0],
	validAccessTypes[1],
	validAccessTypes[2],
	validAccessTypes[3],
}

var formElasticIndex = "forms"

var formElasticType = "form"

var responseElasticIndex = "responses"

var responseElasticType = "response"

var projectElasticIndex = "projects"

var projectElasticType = "project"

var blogElasticIndex = "blogs"

var blogElasticType = "blog"

var formFileIndex = "formfiles"

var responseFileIndex = "responsefiles"

var blogFileIndex = "blogfiles"

var placeholderPath = "/placeholder"

var originalPath = "/original"

var blurPath = "/blur"

var blogSearchFields = []string{
	"title",
	"author",
	"caption",
	"content",
}

var formSearchFields = []string{
	"name",
}

var projectSearchFields = []string{
	"name",
}

var responseSearchFields = []string{
	"items",
}

// all valid file types for attachments
var validContentTypes = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
	"image/svg+xml",
	"video/mpeg",
	"video/webm",
	"video/mp4",
	"video/x-msvideo",
	"application/pdf",
	"text/plain",
	"application/zip",
	"text/csv",
	"application/json",
	"application/ld+json",
	"application/vnd.ms-powerpoint",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation",
	"application/msword",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
}

var haveblur = []string{
	validContentTypes[0],
	validContentTypes[1],
	validContentTypes[2],
}

var progressiveImageSize = 30

var progressiveImageBlurAmount = 20.0

var justDeleteElastic = false

var validUpdateArrayActions = []string{
	"add",
	"remove",
	"move",
	"set",
}

var validUpdateMapActions = []string{
	"add",
	"remove",
	"set",
}

var validFormItemTypes = []string{
	"radio",
	"checkbox",
	"short",
	"text",
	"redgreen",
	"fileupload",
	"fileattachment",
	"media",
}

var validResponseItemTypes = []string{
	validFormItemTypes[0],
	validFormItemTypes[1],
	validFormItemTypes[2],
	validFormItemTypes[4],
	validFormItemTypes[5],
}

var itemTypesRequireOptions = []string{
	validFormItemTypes[0],
	validFormItemTypes[1],
	validFormItemTypes[4],
}

var itemTypesAllowMultipleOptions = []string{
	validFormItemTypes[1],
}

var itemTypesText = []string{
	validFormItemTypes[2],
}

var itemTypesFile = []string{
	validFormItemTypes[5],
}

var autosaveTime = 3 // seconds

var storageAccessTime = 5 // minutes

// more configuration params

var tokenExpiration = 3 // hours
var storageBucketName = "emailhacks"
var cacheTime = time.Duration(3600) * time.Second
