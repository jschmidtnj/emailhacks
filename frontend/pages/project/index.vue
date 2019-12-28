<template>
  <b-container class="mt-4">
    <view-project v-if="projectId" :project-id="projectId" />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import ViewProject from '~/components/secure/project/View.vue'
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
    this.$axios.post('/graphql', {
      query: 'mutation{addProject(name:"",tags:[],categories:[]){id}}'
    })
    .then(res => {
      if (res.status === 200) {
        if (res.data) {
          if (res.data.data && res.data.data.addProject) {
            this.projectId = res.data.data.addProject.id
            history.replaceState({}, null, `/project/${this.projectId}`)
          } else if (res.data.errors) {
            console.error(res.data.errors)
            this.$toasted.global.error({
              message: `found errors: ${JSON.stringify(res.data.errors)}`
            })
          } else {
            this.$toasted.global.error({
              message: 'could not find data or errors'
            })
          }
        } else {
          this.$toasted.global.error({
            message: 'could not get data'
          })
        }
      } else {
        this.$toasted.global.error({
          message: `status code of ${res.status}`
        })
      }
    })
    .catch(err => {
      console.error(err)
      this.$toasted.global.error({
        message: err
      })
    })
  }
})
</script>

<style lang="scss"></style>
