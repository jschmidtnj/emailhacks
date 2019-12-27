<template>
  <b-container class="mt-4">
    <create v-if="id" :id="id" :get-initial-data="true" />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import Create from '~/components/secure/form/Create.vue'
export default Vue.extend({
  name: 'EditForm',
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
