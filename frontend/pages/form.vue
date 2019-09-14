<template>
  <b-container class="mt-4">
    <create />
  </b-container>
</template>

<script lang="ts">
import Vue from 'vue'
import Create from '~/components/secure/form/Create.vue'
export default Vue.extend({
  name: 'Form',
  components: {
    Create
  }
})
</script>

<style lang="scss"></style>
