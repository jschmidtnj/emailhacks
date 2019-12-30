<template>
  <div id="view">
    <b-card no-body>
      <b-card-body>
        <b-card-title>{{ name }}</b-card-title>
      </b-card-body>
    </b-card>
  </div>
</template>

<script lang="js">
import Vue from 'vue'
export default Vue.extend({
  name: 'ViewForm',
  props: {
    formId: {
      type: String,
      default: null
    },
    projectId: {
      type: String,
      default: null
    }
  },
  data() {
    return {
      name: '',
      items: [],
      multiple: false,
      files: []
    }
  },
  mounted() {
    if (this.formId) {
      this.$axios.get('/graphql', {
        params: {
          query: `{form(id:"${this.formId}"){name}}`
        }
      })
      .then(res => {
        if (res.status === 200) {
          if (res.data) {
            if (res.data.data && res.data.data.form) {
              this.name = decodeURIComponent(res.data.data.form.name)
            } else if (res.data.errors) {
              console.error(res.data.errors[0])
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
  }
})
</script>

<style lang="scss"></style>
