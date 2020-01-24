<template>
  <b-container>
    <b-btn @click="copyToClipboard" class="mb-3" variant="primary">
      Copy Email
    </b-btn>
    <view-content :form-id="formId" :project-id="projectId" :preview="true" />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import * as clipboard from 'clipboard-polyfill'
import ViewContent from '~/components/form/View.vue'
export default Vue.extend({
  name: 'SubmitModal',
  components: {
    ViewContent
  },
  props: {
    projectId: {
      type: String,
      default: null
    },
    formId: {
      type: String,
      default: null
    }
  },
  methods: {
    copyToClipboard(evt) {
      evt.preventDefault()
      const dt = new clipboard.DT()
      dt.setData('text/html', '<h1>test data</h1>')
      clipboard.write(dt)
      this.$toasted.global.success({
        message: 'Email copied!'
      })
    }
  }
})
</script>

<style lang="scss"></style>
