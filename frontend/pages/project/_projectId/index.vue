<template>
  <b-container class="mt-4">
    <view-project-data v-if="projectId" :project-id="projectId" />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import ViewProjectData from '~/components/secure/project/View.vue'
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'ViewProject',
  layout: 'secure',
  components: {
    ViewProjectData
  },
  head() {
    const title = 'View Project'
    const description = 'view a project'
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
      projectId: null
    }
  },
  mounted() {
    if (this.$route.params && this.$route.params.projectId) {
      this.projectId = this.$route.params.projectId
    } else {
      this.$nuxt.error({
        statusCode: 404,
        message: 'could not find project id'
      })
    }
  }
})
</script>

<style lang="scss"></style>
