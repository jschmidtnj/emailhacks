<template>
  <b-container class="mt-4">
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
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'Dashboard',
  // @ts-ignore
  layout: 'secure',
  // @ts-ignore
  head() {
    const title = 'Dashboard'
    const description = 'user dashboard'
    const image = `${seo.url}/icon.png`
    return {
      title,
      meta: [
        { property: 'og:title', content: title },
        { property: 'og:description', content: description },
        {
          property: 'og:image',
          content: image
        },
        { name: 'twitter:title', content: title },
        {
          name: 'twitter:description',
          content: description
        },
        {
          name: 'twitter:image',
          content: image
        },
        { hid: 'description', name: 'description', content: description }
      ]
    }
  }
})
</script>

<style lang="scss"></style>
