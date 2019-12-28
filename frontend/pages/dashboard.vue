<template>
  <b-container>
    <b-card>
      <b-list-group>
        <b-list-group-item>Analytics</b-list-group-item>
        <b-list-group-item>Projects</b-list-group-item>
        <b-list-group-item>Project History</b-list-group-item>
        <b-list-group-item>Profile button</b-list-group-item>
      </b-list-group>
    </b-card>
  </b-container>
</template>

<script lang="ts">
import Vue from 'vue'
export default Vue.extend({
  name: 'Dashboard',
  // @ts-ignore
  layout: 'secure'
})
</script>

<style lang="scss"></style>
