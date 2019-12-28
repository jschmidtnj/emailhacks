<template>
  <b-container class="mt-4">
    <view-project v-if="projectId" :project-id="projectId" />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import ViewProject from '~/components/secure/project/View.vue'
export default Vue.extend({
  name: 'ViewProject',
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
    if (this.$route.params && this.$route.params.projectId) {
      this.projectId = this.$route.params.projectId
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
