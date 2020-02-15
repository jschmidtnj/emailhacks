define({ "api": [
  {
    "type": "put",
    "url": "/addResponse",
    "title": "Add response",
    "version": "0.0.1",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "id",
            "description": "<p>Form id for response</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "accessToken",
            "description": "<p>Token for authentication</p>"
          },
          {
            "group": "Parameter",
            "type": "Array",
            "optional": false,
            "field": "items",
            "description": "<p>Item objects</p>"
          },
          {
            "group": "Parameter",
            "type": "Array",
            "optional": false,
            "field": "files",
            "description": "<p>File objects</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "Object",
            "optional": false,
            "field": "data",
            "description": "<p>Response data</p>"
          }
        ]
      }
    },
    "filename": "../api/responseMutations.go",
    "group": "_home_runner_work_emailhacks_emailhacks_api_responseMutations_go",
    "groupTitle": "_home_runner_work_emailhacks_emailhacks_api_responseMutations_go",
    "name": "PutAddresponse",
    "sampleRequest": [
      {
        "url": "https://api.joshuaschmidt.tech/addResponse"
      }
    ]
  },
  {
    "type": "post",
    "url": "/register",
    "title": "User registration",
    "version": "0.0.1",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "email",
            "description": "<p>Registration email</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "password",
            "description": "<p>User password</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "message",
            "description": "<p>Response message</p>"
          }
        ]
      }
    },
    "group": "authentication",
    "filename": "../api/auth.go",
    "groupTitle": "authentication",
    "name": "PostRegister",
    "sampleRequest": [
      {
        "url": "https://api.joshuaschmidt.tech/register"
      }
    ]
  },
  {
    "type": "put",
    "url": "/loginEmailPassword",
    "title": "User login",
    "version": "0.0.1",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "email",
            "description": "<p>User email</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "password",
            "description": "<p>User password</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "token",
            "description": "<p>User token for authenticated requests</p>"
          }
        ]
      }
    },
    "group": "authentication",
    "filename": "../api/auth.go",
    "groupTitle": "authentication",
    "name": "PutLoginemailpassword",
    "sampleRequest": [
      {
        "url": "https://api.joshuaschmidt.tech/loginEmailPassword"
      }
    ]
  },
  {
    "type": "put",
    "url": "/resetPassword",
    "title": "Reset password",
    "version": "0.0.1",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "token",
            "description": "<p>Password reset token</p>"
          },
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "password",
            "description": "<p>New password</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "message",
            "description": "<p>Response message</p>"
          }
        ]
      }
    },
    "group": "authentication",
    "filename": "../api/auth.go",
    "groupTitle": "authentication",
    "name": "PutResetpassword",
    "sampleRequest": [
      {
        "url": "https://api.joshuaschmidt.tech/resetPassword"
      }
    ]
  },
  {
    "type": "put",
    "url": "/sendResetEmail",
    "title": "Send reset email",
    "version": "0.0.1",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "email",
            "description": "<p>User email</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "message",
            "description": "<p>Success message for email sent</p>"
          }
        ]
      }
    },
    "group": "emails",
    "filename": "../api/authEmail.go",
    "groupTitle": "emails",
    "name": "PutSendresetemail",
    "sampleRequest": [
      {
        "url": "https://api.joshuaschmidt.tech/sendResetEmail"
      }
    ]
  },
  {
    "type": "put",
    "url": "/sendTestEmail",
    "title": "Send test email",
    "version": "0.0.1",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "message",
            "description": "<p>Success message for email sent</p>"
          }
        ]
      }
    },
    "group": "emails",
    "filename": "../api/authEmail.go",
    "groupTitle": "emails",
    "name": "PutSendtestemail",
    "sampleRequest": [
      {
        "url": "https://api.joshuaschmidt.tech/sendTestEmail"
      }
    ]
  },
  {
    "type": "get",
    "url": "/",
    "title": "Default rest request",
    "version": "0.0.1",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "message",
            "description": "<p>Message</p>"
          }
        ]
      }
    },
    "group": "misc",
    "filename": "../api/main.go",
    "groupTitle": "misc",
    "name": "Get",
    "sampleRequest": [
      {
        "url": "https://api.joshuaschmidt.tech/"
      }
    ]
  },
  {
    "type": "get",
    "url": "/countBlogs",
    "title": "Count posts for search term",
    "version": "0.0.1",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "searchterm",
            "description": "<p>Search term to count results</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "count",
            "description": "<p>Result count</p>"
          }
        ]
      }
    },
    "group": "misc",
    "filename": "../api/blog.go",
    "groupTitle": "misc",
    "name": "GetCountblogs",
    "sampleRequest": [
      {
        "url": "https://api.joshuaschmidt.tech/countBlogs"
      }
    ]
  },
  {
    "type": "get",
    "url": "/countForms",
    "title": "Count forms for search term",
    "version": "0.0.1",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "searchterm",
            "description": "<p>Search term to count results</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "count",
            "description": "<p>Result count</p>"
          }
        ]
      }
    },
    "group": "misc",
    "filename": "../api/form.go",
    "groupTitle": "misc",
    "name": "GetCountforms",
    "sampleRequest": [
      {
        "url": "https://api.joshuaschmidt.tech/countForms"
      }
    ]
  },
  {
    "type": "get",
    "url": "/countProjects",
    "title": "Count projects for search term",
    "version": "0.0.1",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "searchterm",
            "description": "<p>Search term to count results</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "count",
            "description": "<p>Result count</p>"
          }
        ]
      }
    },
    "group": "misc",
    "filename": "../api/project.go",
    "groupTitle": "misc",
    "name": "GetCountprojects",
    "sampleRequest": [
      {
        "url": "https://api.joshuaschmidt.tech/countProjects"
      }
    ]
  },
  {
    "type": "get",
    "url": "/countResponses",
    "title": "Count responses for search term",
    "version": "0.0.1",
    "parameter": {
      "fields": {
        "Parameter": [
          {
            "group": "Parameter",
            "type": "String",
            "optional": false,
            "field": "searchterm",
            "description": "<p>Search term to count results</p>"
          }
        ]
      }
    },
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "count",
            "description": "<p>Result count</p>"
          }
        ]
      }
    },
    "group": "misc",
    "filename": "../api/response.go",
    "groupTitle": "misc",
    "name": "GetCountresponses",
    "sampleRequest": [
      {
        "url": "https://api.joshuaschmidt.tech/countResponses"
      }
    ]
  },
  {
    "type": "get",
    "url": "/hello",
    "title": "Test rest request",
    "version": "0.0.1",
    "success": {
      "fields": {
        "Success 200": [
          {
            "group": "Success 200",
            "type": "String",
            "optional": false,
            "field": "message",
            "description": "<p>Hello message</p>"
          }
        ]
      }
    },
    "group": "misc",
    "filename": "../api/main.go",
    "groupTitle": "misc",
    "name": "GetHello",
    "sampleRequest": [
      {
        "url": "https://api.joshuaschmidt.tech/hello"
      }
    ]
  }
] });
