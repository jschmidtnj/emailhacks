<template>
  <b-container>
    <b-jumbotron :header="$t('welcome')" class="mt-4" />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'Home',
  data() {
    return {}
  },
  i18n: {
    messages: {
      'en': require('~/locales/index/en.js')
    }
  },
  // @ts-ignore
  head() {
    const title = 'Home'
    const description = 'mail pear - email reimagined'
    const image = `${seo.url}/icon.png`
    return {
      title,
      meta: [
        { property: 'og:title', content: title },
        { property: 'og:description', content: description },
        {
          property: 'og:image',
          content: image
        },
        { name: 'twitter:title', content: title },
        {
          name: 'twitter:description',
          content: description
        },
        {
          name: 'twitter:image',
          content: image
        },
        { hid: 'description', name: 'description', content: description }
      ],
      __dangerouslyDisableSanitizers: ['script'],
      script: [
        {
          innerHTML: JSON.stringify({
            '@context': 'http://schema.org',
            '@type': 'WebPage',
            name: title,
            description
          }),
          type: 'application/ld+json'
        }
      ]
    }
  }
})
</script>

<style lang="scss"></style>
