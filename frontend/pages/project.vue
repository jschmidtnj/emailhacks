<template>
  <b-container class="mt-4">
    <page-loading :loading="true" />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import PageLoading from '~/components/PageLoading.vue'
import { defaultItemName } from '~/assets/config'
const seo = JSON.parse(process.env.seoconfig)
// create a new project
export default Vue.extend({
  name: 'NewProject',
  layout: 'secure',
  components: {
    PageLoading
  },
  head() {
    const title = 'New Project'
    const description = 'create a new project'
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
    this.$apollo
      .mutate({
        mutation: gql`
          mutation addProject(
            $name: String!
            $tags: [String!]!
            $categories: [String!]!
          ) {
            addProject(name: $name, tags: $tags, categories: $categories) {
              id
            }
          }
        `,
        variables: { name: defaultItemName, categories: [], tags: [] }
      })
      .then(({ data }) => {
        if (data.addProject && data.addProject.id) {
          this.$store.commit('project/setProjectId', data.addProject.id)
          this.$store.commit('project/setProjectName', defaultItemName)
          this.$bvToast.toast('added project', {
            variant: 'success',
            title: 'Success'
          })
          this.$router.push({
            path: '/dashboard'
          })
        } else {
          this.$bvToast.toast('cannot find project id', {
            variant: 'danger',
            title: 'Error'
          })
        }
      })
      .catch((err) => {
        this.$bvToast.toast(`found error: ${err.message}`, {
          variant: 'danger',
          title: 'Error'
        })
      })
  }
})
</script>

<style lang="scss"></style>
