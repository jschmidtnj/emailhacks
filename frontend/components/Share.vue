<template>
  <b-modal id="share-modal" v-if="shareType" title="BootstrapVue">
    <p class="my-4">Share</p>
    <vue-tags-input
      v-model="tag"
      :tags="tags"
      @tags-changed="(newTags) => (tags = newTags)"
    />
  </b-modal>
</template>

<script lang="js">
import Vue from 'vue'
import VueTagsInput from '@johmun/vue-tags-input'
// TODO - get access and have the sharing endpoints enabled
// check if user exists before sharing with them
export default Vue.extend({
  name: 'Share',
  components: {
    VueTagsInput
  },
  props: {
    access: {
      type: Array,
      default: null
    }
  },
  data() {
    return {
      tags: [],
      tag: ''
    }
  }
})
</script>

<style lang="scss"></style>
