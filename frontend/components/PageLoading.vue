<template>
  <atom-spinner
    v-if="loading"
    :animation-duration="1000"
    :size="60"
    :color="'#ff1d5e'"
    class="centered"
  />
</template>

<script lang="js">
import Vue from 'vue'
import { AtomSpinner } from 'epic-spinners/dist/lib/epic-spinners.min.js'
export default Vue.extend({
  name: 'PageLoading',
  components: {
    AtomSpinner
  },
  props: {
    loading: {
      type: Boolean,
      default: false
    }
  },
  methods: {
    start() {
      this.loading = true
    },
    finish() {
      this.loading = false
    }
  }
})
</script>

<style lang="scss">
.centered {
  display: block;
  margin-left: auto;
  margin-right: auto;
}
</style>
