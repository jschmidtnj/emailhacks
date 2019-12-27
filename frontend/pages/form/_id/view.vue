<template>
  <b-container class="mt-4">
    <view-content v-if="id" />
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
      id: null
    }
  },
  mounted() {
    if (this.$route.params && this.$route.params.id) {
      this.id = this.$route.params.id
    } else {
      this.$nuxt.error({
        statusCode: 404,
        message: 'could not find form id'
      })
    }
  }
})
</script>

<style lang="scss"></style>
