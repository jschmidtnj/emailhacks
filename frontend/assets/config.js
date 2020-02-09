export const codes = {
  error: 400,
  success: 200,
  warning: 300,
  unauthorized: 403
}

export const cloudStorageURLs = {
  static: 'https://storage.googleapis.com/emailhacks'
}

export const staticStorageIndexes = {
  blogfiles: 'blogfiles',
  formfiles: 'formfiles',
  placeholder: 'placeholder'
}

export const validTypes = ['project', 'form']

// autosave this quickly
export const autosaveInterval = 1 * 1000

// periodically check if logged in
export const checkLoggedInInterval = 5 * 60 * 1000

// default item name when first created
export const defaultItemName = 'Untitled'

export const defaultCountry = 'US'

export const defaultCurrency = 'USD'

export const adminTypes = ['admin', 'super']

export const plans = ['free', 'business', 'enterprise']

export const toasts = {
  position: 'top-right',
  duration: 2000,
  theme: 'bubble'
}

export const regex = {
  password: /^$|^(?=.*[A-Za-z])(?=.*\d)(?=.*[@$!%*#?&])[A-Za-z\d@$!%*#?&]{8,}$/,
  hexcode: /(^#[0-9A-F]{6}$)|(^#[0-9A-F]{3}$)/i,
  phone: /^[+]*[(]{0,1}[0-9]{1,4}[)]{0,1}[-\s\\./0-9]*$/
}

export const oauthConfig = {
  google: {
    oauth2Endpoint: 'https://accounts.google.com/o/oauth2/v2/auth',
    scope: ['profile', 'email'].join(' ')
  }
}

export const paths = {
  placeholder: '/placeholder',
  original: '/original',
  blur: '/blur'
}

export const options = {
  categoryOptions: ['technology', 'webdesign'],
  tagOptions: ['vue', 'nuxt']
}

export const defaultColor = '#194d332B'

export const validfiles = [
  'image/jpeg',
  'image/png',
  'image/gif',
  'image/svg+xml',
  'video/mpeg',
  'video/mp4',
  'video/webm',
  'video/x-msvideo',
  'application/pdf',
  'text/plain',
  'application/zip',
  'text/csv',
  'application/json',
  'application/ld+json',
  'application/vnd.ms-powerpoint',
  'application/vnd.openxmlformats-officedocument.presentationml.presentation',
  'application/msword',
  'application/vnd.openxmlformats-officedocument.wordprocessingml.document'
]

export const validimages = [validfiles[0], validfiles[1], validfiles[3]]

export const validDisplayFiles = [
  ...validimages,
  validfiles[2],
  validfiles[4],
  validfiles[5],
  validfiles[6]
]

export const noneAccessType = 'none'
