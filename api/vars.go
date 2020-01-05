package main

import ()

var mainDatabase = "website"

var userMongoName = "users"

var formMongoName = "forms"

var projectMongoName = "projects"

var blogMongoName = "blogs"

var shortLinkMongoName = "shortlink"

var tokenKey = "token"

var miscKey = "misc"

var dataKey = "data"

var graphiQL = false

var graphqlPlayground = true

var sitemapTimeFormat = "2006-01-02T15:04:05Z07:00"

var dateFormat = "Mon Jan _2 15:04:05 2006"

var hexRegex = "(^#[0-9A-F]{6}$)|(^#[0-9A-F]{8}$)"

var formType = "form"

var projectType = "project"

var blogType = "blog"

var validTypes = []string{
	formType,
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
	"none",
}

var noAccessLevel = validAccessTypes[3]

var editAccessLevel = []string{
	validAccessTypes[0],
	validAccessTypes[1],
}

var viewAccessLevel = []string{
	validAccessTypes[0],
	validAccessTypes[1],
	validAccessTypes[2],
}

var formElasticIndex = "forms"

var formElasticType = "form"

var projectElasticIndex = "projects"

var projectElasticType = "project"

var blogElasticIndex = "blogs"

var blogElasticType = "blog"

var formFileIndex = "formfiles"

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

var autosaveTime = 3 // seconds
