<template>
  <b-container class="mt-4">
    <view-content
      v-if="projectId && formId && responseId"
      :form-id="formId"
      :project-id="projectId"
      :response-id="responseId"
    />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import ViewContent from '~/components/form/View.vue'
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'Response',
  components: {
    ViewContent
  },
  head() {
    const title = 'Response'
    const description = 'view / edit a response'
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
      projectId: null,
      formId: null
    }
  },
  mounted() {
    if (this.$route.params && this.$route.params.projectId
      && this.$route.params.formId && this.$route.params.responseId) {
      this.projectId = this.$route.params.projectId
      this.formId = this.$route.params.formId
      this.responseId = this.$route.params.responseId
    } else {
      this.$nuxt.error({
        statusCode: 404,
        message: 'could not find form id or project id or response id'
      })
    }
  }
})
</script>

<style lang="scss"></style>
