<template>
  <b-container class="mt-4">
    <view-project
      v-if="projectId"
      :project-id="projectId"
      :get-initial-data="false"
    />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import ViewProject from '~/components/secure/project/View.vue'
import { defaultItemName } from '~/assets/config'
export default Vue.extend({
  name: 'NewProject',
  layout: 'secure',
  components: {
    ViewProject
  },
  data() {
    return {
      projectId: null
    }
  },
  mounted() {
    this.$apollo.mutate({mutation: gql`
      mutation addProject($name: String!, $tags: [String!]!, $categories: [String!]!)
      {addProject(name: $name, tags: $tags, categories: $categories){id} }
      `, variables: {name: defaultItemName, categories: [], tags: []}})
      .then(({ data }) => {
        this.projectId = data.addProject.id
        history.replaceState({}, null, `/project/${this.projectId}`)
      }).catch(err => {
        console.error(err)
        this.$toasted.global.error({
          message: `found error: ${err.message}`
        })
      })
  }
})
</script>

<style lang="scss"></style>
