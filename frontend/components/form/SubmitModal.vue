<template>
  <b-modal ref="submit-content-modal" hide-footer size="xl" title="Preview">
    <view-content :form-id="formId" :project-id="projectId" :preview="true" />
  </b-modal>
</template>

<script lang="js">
import Vue from 'vue'
import ViewContent from '~/components/form/View.vue'
// TODO - get the email template to copy paste
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
  }
})
</script>

<style lang="scss"></style>
