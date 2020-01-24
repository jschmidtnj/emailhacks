<template>
  <b-container>
    <main-checkout class="mt-4" />
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import MainCheckout from '~/components/Checkout.vue'
export default Vue.extend({
  name: 'Checkout',
  layout: 'secure',
  components: {
    MainCheckout
  },
  data() {
    return {}
  }
})
</script>

<style lang="scss"></style>
