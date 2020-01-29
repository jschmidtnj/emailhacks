<template>
  <b-modal ref="send-content-modal" hide-footer size="xl" title="Preview">
    <b-container>
      <b-form-group>
        <label class="form-required">Email Address</label>
        <span>
          <b-form-input
            id="email"
            v-model="email"
            :state="!$v.email.$invalid"
            type="text"
            class="form-control"
            placeholder="email"
            aria-describedby="emailfeedback"
          />
        </span>
        <b-form-invalid-feedback id="emailfeedback" :state="!$v.email.$invalid">
          <div v-if="!$v.email.required">email is required</div>
          <div v-else-if="!$v.email.email">email is invalid</div>
        </b-form-invalid-feedback>
      </b-form-group>
      <b-btn
        :disabled="$v.email.$invalid"
        @click="copyToClipboard"
        class="mb-3"
        variant="primary"
      >
        Copy Email
      </b-btn>
      <view-content :form-id="formId" :preview="true" />
    </b-container>
  </b-modal>
</template>

<script lang="js">
import Vue from 'vue'
import { validationMixin } from 'vuelidate'
import { required, email } from 'vuelidate/lib/validators'
import * as clipboard from 'clipboard-polyfill'
import ViewContent from '~/components/form/View.vue'
export default Vue.extend({
  name: 'SendModal',
  components: {
    ViewContent
  },
  mixins: [validationMixin],
  props: {
    formId: {
      type: String,
      default: null
    }
  },
  data() {
    return {
      email: ''
    }
  },
  validations: {
    email: {
      required,
      email
    }
  },
  methods: {
    show() {
      if (this.$refs['send-content-modal']) {
        this.$refs['send-content-modal'].show()
      } else {
        this.$bvToast.toast('cannot find send modal', {
          variant: 'danger',
          title: 'Error'
        })
      }
    },
    copyToClipboard(evt) {
      evt.preventDefault()
      const dt = new clipboard.DT()
      dt.setData('text/html', '<h1>test data</h1>')
      clipboard.write(dt)
      this.$bvToast.toast('Email copied!', {
        variant: 'success',
        title: 'Success'
      })
    }
  }
})
</script>

<style lang="scss"></style>
