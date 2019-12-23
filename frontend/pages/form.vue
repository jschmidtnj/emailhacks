<template>
  <b-container class="mt-4">
    <create />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import Create from '~/components/secure/form/Create.vue'
export default Vue.extend({
  name: 'Form',
  components: {
    Create
  },
  layout: 'secure'
})
</script>

<style lang="scss"></style>
