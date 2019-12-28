<template>
  <b-container class="mt-4">
    <create
      v-if="formId && projectId"
      :form-id="formId"
      :project-id="projectId"
      :get-initial-data="false"
    />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import Create from '~/components/secure/form/Create.vue'
export default Vue.extend({
  name: 'NewForm',
  layout: 'secure',
  components: {
    Create
  },
  data() {
    return {
      projectId: null,
      formId: null
    }
  },
  mounted() {
    if (this.$route.params && this.$route.params.projectId) {
      this.projectId = this.$route.params.projectId
      this.$axios.post('/graphql', {
        query: `mutation{addForm(title:"",items:[],multiple:false,tags:[],categories:[],files:[]){id}}`
      })
      .then(res => {
        if (res.status === 200) {
          if (res.data) {
            if (res.data.data && res.data.data.addForm) {
              this.formId = res.data.data.addForm.id
              history.replaceState({}, null, `/project/${this.projectId}/form/${this.formId}/edit`)
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
