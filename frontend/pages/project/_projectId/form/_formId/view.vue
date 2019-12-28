<template>
  <b-container class="mt-4">
    <view-content
      v-if="projectId && formId"
      :form-id="formId"
      :project-id="projectId"
    />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import ViewContent from '~/components/secure/form/View.vue'
export default Vue.extend({
  name: 'ViewForm',
  components: {
    ViewContent
  },
  data() {
    return {
      projectId: null,
      formId: null
    }
  },
  mounted() {
    if (this.$route.params && this.$route.params.projectId && this.$route.params.formId) {
      this.projectId = this.$route.params.projectId
      this.formId = this.$route.params.formId
    } else {
      this.$nuxt.error({
        statusCode: 404,
        message: 'could not find form id or project id'
      })
    }
  }
})
</script>

<style lang="scss"></style>
