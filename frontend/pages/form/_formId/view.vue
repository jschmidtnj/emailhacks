<template>
  <b-container class="mt-4">
    <view-content v-if="formId" :form-id="formId" :access-token="accessToken" />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import ViewContent from '~/components/form/View.vue'
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'ViewForm',
  components: {
    ViewContent
  },
  head() {
    const title = 'View Form'
    const description = 'view a form'
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
      ]
    }
  },
  data() {
    return {
      formId: null,
      accessToken: null
    }
  },
  mounted() {
    if (this.$route.query.accessToken) {
      this.accessToken = this.$route.query.accessToken
    }
    if (this.$route.params && this.$route.params.formId) {
      this.formId = this.$route.params.formId
    } else {
      this.$nuxt.error({
        statusCode: 404,
        message: 'could not find form id or project id'
      })
    }
  }
})
</script>

<style lang="scss"></style>
