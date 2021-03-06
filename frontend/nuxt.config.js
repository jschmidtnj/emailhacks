const pkg = require('./package')
require('dotenv').config()
const seodata = JSON.parse(process.env.SEOCONFIG)
const apiurl = process.env.APIURL
const useSecure = process.env.USESECURE !== 'false'
const recaptchasitekey = process.env.RECAPTCHASITEKEY
const fullApiUrl = `${useSecure ? 'https' : 'http'}://${apiurl}`

const name = 'Mail Pear'

module.exports = {
  mode: 'spa',

  globalName: name,

  server: {
    port: 8080 // default: 3000
  },

  env: {
    mapsautoapikey: process.env.MAPSAUTOAPIKEY,
    seoconfig: process.env.SEOCONFIG,
    githuburl: pkg.repository.url,
    authconfig: process.env.AUTHCONFIG,
    apiurl: fullApiUrl,
    stripeconfig: process.env.STRIPECONFIG,
    recaptchasitekey,
    shortlinkurl: process.env.SHORTLINKURL
  },

  /*
   ** Headers of the page
   */
  head: {
    titleTemplate: `%s - ${name}`,
    meta: [
      { charset: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      // OpenGraph Data
      { property: 'og:site_name', content: name },
      // The list of types is available here: http://ogp.me/#types
      { property: 'og:type', content: 'website' },
      // Twitter card
      { name: 'twitter:card', content: 'summary_large_image' },
      {
        name: 'twitter:site',
        content: `@${seodata.twitterhandle}`
      },
      { name: 'twitter:creator', content: `@${seodata.twitterhandle}` }
    ],
    link: [{ rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }],
    __dangerouslyDisableSanitizers: ['script'],
    script: [
      {
        src: 'https://js.stripe.com/v3/',
        defer: true
      },
      {
        src: `https://www.google.com/recaptcha/api.js?render=${recaptchasitekey}`,
        defer: true
      },
      {
        innerHTML: JSON.stringify({
          '@context': 'https://schema.org',
          '@type': 'Organization',
          name,
          url: seodata.url,
          logo: `${seodata.url}/icon.png`,
          contactPoint: {
            '@type': 'ContactPoint',
            email: seodata.email,
            contactType: 'technical support',
            url: `${seodata.url}/about`
          },
          sameAs: [
            `https://twitter.com/${seodata.twitterhandle}`,
            seodata.facebook,
            seodata.linkedin,
            seodata.github
          ]
        }),
        type: 'application/ld+json'
      },
      {
        innerHTML: JSON.stringify({
          '@context': 'http://schema.org',
          '@type': 'WebSite',
          url: seodata.url,
          potentialAction: {
            '@type': 'SearchAction',
            target: `${seodata.url}/forms?phrase={query}`,
            'query-input': 'required name=query'
          }
        }),
        type: 'application/ld+json'
      }
    ]
  },
  /*
   ** Customize the progress-bar color
   */
  loading: { color: '#fff' },
  /*
   ** Customize loading icon
   */
  loadingIndicator: {
    name: 'pulse',
    color: '#3B8070',
    background: 'white'
  },
  /*
   ** Global CSS
   */
  css: [],
  /*
   ** fix vue meta
   */
  vueMeta: {
    debounceWait: 50
  },
  /*
   ** Plugins to load before mounting the App
   */
  plugins: [
    { src: '~/plugins/font-awesome', ssr: false },
    { src: '~/plugins/vuelidate', ssr: false },
    { src: '~/plugins/vuex-persist', ssr: false },
    { src: '~/plugins/axios', ssr: false },
    { src: '~/plugins/donut', ssr: false },
    { src: '~/plugins/scroll', ssr: false },
    { src: '~/plugins/scroll-reveal', ssr: false },
    { src: '~/plugins/touch', ssr: false }
  ],
  /*
   ** Nuxt.js dev-modules
   */
  buildModules: [
    // Doc: https://github.com/nuxt-community/eslint-module
    '@nuxtjs/eslint-module',
    '@nuxt/typescript-build'
  ],
  /*
   ** Nuxt.js modules
   */
  modules: [
    // Doc: https://bootstrap-vue.js.org
    'bootstrap-vue/nuxt',
    // Doc: https://axios.nuxtjs.org/usage
    '@nuxtjs/axios',
    '@nuxtjs/pwa',
    // Doc: https://github.com/nuxt-community/dotenv-module
    '@nuxtjs/dotenv',
    '@nuxtjs/sitemap',
    '@nuxtjs/style-resources',
    'nuxt-webfontloader',
    '@nuxtjs/google-analytics',
    '@nuxtjs/apollo',
    'nuxt-i18n'
  ],
  /*
   ** i8n config
   */
  i18n: {
    seo: false, // defined in heads in layout
    baseUrl: seodata.url,
    locales: [
      {
        code: 'en',
        iso: 'en-US'
      },
      {
        code: 'es',
        iso: 'es-ES'
      }
    ],
    defaultLocale: 'en',
    vueI18nLoader: true,
    vueI18n: {
      fallbackLocale: 'en',
      messages: {
        en: {},
        es: {}
      }
    }
  },
  /*
   ** apollo config
   */
  apollo: {
    tokenName: 'mail-pear-apollo-token',
    cookieAttributes: {
      expires: null, // only for this session
      // domain: seodata.url, // defaults to domain where it was created
      secure: useSecure
    },
    clientConfigs: {
      default: {
        httpEndpoint: fullApiUrl + '/graphql',
        // Use websockets for everything (no HTTP)
        // You need to pass a `wsEndpoint` for this to work
        websocketsOnly: false,
        tokenName: 'mail-pear-apollo-token',
        wsEndpoint: `${useSecure ? 'wss' : 'ws'}://${apiurl}/subscriptions`
      }
    }
  },
  /*
   ** scss global config
   */
  styleResources: {
    scss: ['~assets/styles/global.scss']
  },
  /*
   ** google web fonts
   */
  webfontloader: {
    google: {
      families: ['Roboto']
    }
  },
  /*
   ** google analytics config
   */
  googleAnalytics: {
    id: seodata.googleanalyticstrackingid
  },
  /*
   ** generate config
   */
  generate: {
    fallback: '404.html'
  },
  /*
   ** Axios module configuration
   ** See https://axios.nuxtjs.org/options
   */
  axios: {
    baseURL: fullApiUrl
  },
  /*
   ** Sitemap config
   */
  sitemap: {
    hostname: seodata.url,
    path: '/sitemap-main.xml',
    gzip: false,
    exclude: ['/admin', '/admin/**', '/callback'],
    defaults: {
      changefreq: 'daily',
      priority: 1,
      lastmod: new Date(),
      lastmodrealtime: true
    }
  },
  /*
   ** Build configuration
   */
  build: {
    /*
     ** You can extend webpack config here
     */
    extend(config, ctx) {}
  }
}
