define({ "api": [
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
    "filename": "../api/email.go",
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
    "filename": "../api/email.go",
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
