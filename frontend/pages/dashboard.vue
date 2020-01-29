<template>
  <b-container class="mt-4">
    <view-project-data v-if="$store.state.project.project" />
    <nuxt-link to="/project" class="btn btn-primary btn-sm no-underline mt-4">
      Create New Project
    </nuxt-link>
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import ViewProjectData from '~/components/project/View.vue'
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'Dashboard',
  // @ts-ignore
  layout: 'secure',
  components: {
    ViewProjectData
  },
  // @ts-ignore
  head() {
    const title = 'Dashboard'
    const description = 'project dashboard'
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
    return {}
  },
  mounted() {
    if (this.$route.query && this.$route.query.project) {
      this.$store.commit('project/setProject', this.$route.query.project)
    }
  }
})
</script>

<style lang="scss"></style>
