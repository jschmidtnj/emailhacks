<template>
  <div>
    <h3>Add:</h3>
    <add-organize />
  </div>
</template>

<script lang="js">
import Vue from 'vue'
import AddOrganize from '~/components/AddOrganize.vue'
// TODO - add and remove organization
export default Vue.extend({
  name: 'OrganizeEdit',
  components: {
    AddOrganize
  }
})
</script>
