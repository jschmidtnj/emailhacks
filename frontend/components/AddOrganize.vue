<template>
  <b-component>
    <h3>new {{ addType }}</h3>
    <b-row>
      <b-form-input
        id="organizeInput"
        v-model="organizeName"
        :disabled="true"
        type="text"
      />
      <b-dropdown v-model="addType" @change="changedAddType" :text="addType">
        <b-dropdown-item>Tag</b-dropdown-item>
        <b-dropdown-item>Category</b-dropdown-item>
      </b-dropdown>
      <b-btn
        @click="submit"
        :disabled="$v.organizeName.$invalid"
        class="mb-3"
        variant="primary"
      >
        Submit
      </b-btn>
    </b-row>
  </b-component>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import { validationMixin } from 'vuelidate'
import { required, minLength } from 'vuelidate/lib/validators'
const addTypes = ['tag', 'category']
export default Vue.extend({
  name: 'AddOrganize',
  mixins: [validationMixin],
  data() {
    return {
      addType: addTypes[0],
      organizeName: '',
      color: ''
    }
  },
  // @ts-ignore
  validations: {
    organizeName: {
      required,
      minLength: minLength(3)
    }
  },
  methods: {
    changedAddType(evt) {
      evt.preventDefault()
      this.addType = this.addType === addTypes[0] ? addTypes[1] : addTypes[0]
    },
    submit(evt) {
      evt.preventDefault()
      if (!this.$store.state.auth.user) {
        this.$bvToast.toast('cannot find user data', {
          variant: 'danger',
          title: 'Error'
        })
        return
      }
      const currentOrganizeData = this.addType === addTypes[0]
        ? this.$store.state.auth.user.tags : this.$store.state.auth.user.categories
      if (!currentOrganizeData) {
        this.$bvToast.toast('cannot find current organize data', {
          variant: 'danger',
          title: 'Error'
        })
        return
      }
      if (currentOrganizeData.includes(this.organizeName)) {
        this.$bvToast.toast('organization already ', {
          variant: 'danger',
          title: 'Error'
        })
        return
      }
      this.$apollo.mutate({mutation: gql`
        mutation addOrganization($type: String!, $name: String!, $color: String!) {
          addOrganization(type: $type, name: $name, color: $color) {
            id
          }
        }
        `, variables: {
          color: this.color,
          type: this.addType,
          name: this.organizeName
        }})
        .then(({ data }) => {
          currentOrganizeData.push({
            name: this.organizeName,
            color: this.color
          })
          this.organizeName = ''
          this.$bvToast.toast('added organization', {
            variant: 'success',
            title: 'Success'
          })
        }).catch(err => {
          this.$bvToast.toast(`found error: ${err.message}`, {
            variant: 'danger',
            title: 'Error'
          })
        })
    }
  }
})
</script>

<style lang="scss"></style>
