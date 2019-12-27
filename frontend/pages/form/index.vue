<template>
  <b-container class="mt-4">
    <create v-if="id" :id="id" :get-initial-data="false" />
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
      id: null
    }
  },
  mounted() {
    this.$axios.post('/graphql', {
      query: 'mutation{addForm(subject:"",recipient:"",items:[],multiple:false,tags:[],categories:[],files:[]){id}}'
    })
    .then(res => {
      if (res.status === 200) {
        if (res.data) {
          if (res.data.data && res.data.data.addForm) {
            this.id = res.data.data.addForm.id
            history.replaceState({}, null, `/form/${this.id}/edit`)
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
