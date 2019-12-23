<template>
  <div id="view">
    <b-card no-body>
      <b-card-body>
        <b-card-title>From Title</b-card-title>
        <b-card-sub-title class="mb-2">Form Description</b-card-sub-title>
      </b-card-body>
    </b-card>
  </div>
</template>

<script lang="js">
import Vue from 'vue'
export default Vue.extend({
  name: 'View',
  data() {
    return {
      name: '',
      description: '',
      questions: []
    }
  }
})
</script>

<style lang="scss"></style>
