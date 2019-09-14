const pkg = require('./package')
require('dotenv').config()
const seodata = JSON.parse(process.env.SEOCONFIG)
const apiurl = process.env.APIURL
const recaptchasitekey = process.env.RECAPTCHASITEKEY

module.exports = {
  mode: 'spa',

  globalName: pkg.author,

  env: {
    seoconfig: process.env.SEOCONFIG,
    githuburl: pkg.repository.url,
    authconfig: process.env.AUTHCONFIG,
    apiurl: apiurl,
    recaptchasitekey: recaptchasitekey
  },

  /*
   ** Headers of the page
   */
  head: {
    titleTemplate: `%s - ${pkg.author}`,
    meta: [
      { charset: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      // OpenGraph Data
      { property: 'og:site_name', content: pkg.author },
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
        innerHTML: JSON.stringify({
          '@context': 'https://schema.org',
          '@type': 'Organization',
          name: pkg.author,
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
   ** Router config
   */
  router: {},

  /*
   ** Customize the progress-bar color
   */
  loading: { color: '#fff' },

  /*
   ** Global CSS
   */
  css: [],

  /*
   ** Plugins to load before mounting the App
   */
  plugins: [
    { src: '~/plugins/font-awesome', ssr: false },
    { src: '~/plugins/vuelidate', ssr: false },
    { src: '~/plugins/vuex-persist', ssr: false },
    { src: '~/plugins/axios', ssr: false },
    { src: '~/plugins/toast', ssr: false },
    { src: '~/plugins/select', ssr: false },
    { src: '~/plugins/recaptcha', ssr: false },
    { src: '~/plugins/scroll-reveal', ssr: false },
    { src: '~/plugins/pdf', ssr: false },
    { src: '~/plugins/draggable', ssr: false }
  ],

  /*
   ** Nuxt.js modules
   */
  modules: [
    // Doc: https://axios.nuxtjs.org/usage
    '@nuxtjs/axios',
    // Doc: https://bootstrap-vue.js.org/docs/
    'bootstrap-vue/nuxt',
    '@nuxtjs/pwa',
    '@nuxtjs/style-resources',
    '@nuxtjs/dotenv',
    '@nuxtjs/sitemap',
    'nuxt-webfontloader',
    '@nuxtjs/google-analytics'
  ],

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
   ** Axios module configuration
   */
  axios: {
    // See https://github.com/nuxt-community/axios-module#options
    baseURL: apiurl
  },

  /*
   ** babel config
   */
  babel: {
    presets: ['es2015', 'stage-0'],
    plugins: [
      [
        'transform-runtime',
        {
          polyfill: true,
          regenerator: true
        }
      ]
    ]
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

  extensions: ['js', 'ts'],

  /*
   ** Build configuration
   */
  build: {
    // put CSS in files instead of JS bundles
    extractCSS: true,
    /*
     ** You can extend webpack config here
     */
    extend(config, ctx) {
      config.output.globalObject = 'this'
      // Run ESLint on save
      if (ctx.isDev && ctx.isClient) {
        config.module.rules.push({
          enforce: 'pre',
          test: /\.(js|vue)$/,
          loader: 'eslint-loader',
          exclude: /(node_modules)/
        })
      }
    }
  }
}
