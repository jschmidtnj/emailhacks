<template>
  <b-container class="mt-4">
    <b-card title="Login" footer-tag="footer">
      <b-form @submit="loginlocal">
        <b-form-group
          id="email-address-group"
          label="Email address:"
          label-for="email-address"
          description="Your email is safe with us"
        >
          <b-form-input
            id="email-address"
            v-model="form.email"
            :state="!$v.form.email.$invalid"
            type="text"
            placeholder="Enter email"
            aria-describedby="emailfeedback"
          />
          <b-form-invalid-feedback
            id="emailfeedback"
            :state="!$v.form.email.$invalid"
          >
            <div v-if="!$v.form.email.required">
              email is required
            </div>
            <div v-else-if="!$v.form.email.email">
              email is invalid
            </div>
          </b-form-invalid-feedback>
        </b-form-group>
        <b-form-group
          id="password-group"
          label="Password:"
          label-for="password"
          description="Password must be at least 8 characters long, with a number, capital letter and special character"
        >
          <b-form-input
            id="password"
            v-model="form.password"
            :state="!$v.form.password.$invalid"
            type="password"
            placeholder="Enter password"
            aria-describedby="passwordfeedback"
          />
          <b-form-invalid-feedback
            id="passwordfeedback"
            :state="!$v.form.password.$invalid"
          >
            <div v-if="!$v.form.password.required">
              password is required
            </div>
            <div v-else-if="!$v.form.password.validPassword">
              password is invalid
            </div>
          </b-form-invalid-feedback>
        </b-form-group>
        <b-link href="/reset" class="card-link">
          reset password
        </b-link>
        <br />
        <b-button
          :disabled="$v.form.$invalid"
          variant="primary"
          type="submit"
          class="mt-4"
        >
          Submit
        </b-button>
      </b-form>
      <p slot="footer">
        By clicking submit you aggree to the
        <a href="/privacy">privacy policy</a>. This site is protected by
        reCAPTCHA and the Google
        <a href="https://policies.google.com/privacy">Privacy Policy</a> and
        <a href="https://policies.google.com/terms">Terms of Service</a> apply.
      </p>
    </b-card>
  </b-container>
</template>

<script lang="js">
import Vue from 'vue'
import { validationMixin } from 'vuelidate'
import { required, email } from 'vuelidate/lib/validators'
import { regex } from '~/assets/config'
const validPassword = (val) => regex.password.test(val)
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
export default Vue.extend({
  name: 'Login',
  mixins: [validationMixin],
  middleware: 'loginRedirect',
  data() {
    return {
      redirect_uri: null,
      form: {
        email: '',
        password: ''
      }
    }
  },
  // @ts-ignore
  head() {
    const title = 'Login'
    const description = 'login to your account'
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
  },
  // @ts-ignore
  validations: {
    form: {
      email: {
        required,
        email
      },
      password: {
        required,
        validPassword
      }
    }
  },
  mounted() {
    if (
      this.$route.query &&
      this.$route.query.verify &&
      this.$route.query.token
    ) {
      this.$axios
        .post('/verify', {
          token: this.$route.query.token
        })
        .then((res) => {
          if (res.status === 200) {
            if (res.data) {
              let message = 'email verified. you can now log in'
              if (res.data.message) {
                message = res.data.message
              }
              this.$bvToast.toast(message, {
                variant: 'success',
                title: 'Success'
              })
            } else {
              this.$bvToast.toast('could not get data', {
                variant: 'danger',
                title: 'Error'
              })
            }
          } else if (res.data && res.data.message) {
            this.$bvToast.toast(res.data.message, {
              variant: 'danger',
              title: 'Error'
            })
          } else {
            this.$bvToast.toast(`status code of ${res.status}`, {
              variant: 'danger',
              title: 'Error'
            })
          }
        })
        .catch((err) => {
          let message = `got error: ${err}`
          if (err.response && err.response.data) {
            message = err.response.data.message
          }
          this.$bvToast.toast(message, {
            variant: 'danger',
            title: 'Error'
          })
        })
    }
    if (this.$route.query && this.$route.query.redirect_uri) {
      this.redirect_uri = this.$route.query.redirect_uri
    }
  },
  methods: {
    loginlocal(evt) {
      evt.preventDefault()
      if (this.redirect_uri) {
        if (this.redirect_uri.indexOf('/') === 0) {
          this.$router.push({
            path: this.redirect_uri
          })
        } else {
          this.$bvToast.toast('invalid url redirect', {
            variant: 'danger',
            title: 'Error'
          })
        }
      }
      this.$recaptcha('login')
        .then((recaptchatoken) => {
          /* eslint-disable */
          console.log(`got recaptcha token ${recaptchatoken}`)
          this.$store
            .dispatch('auth/loginLocal', {
              email: this.form.email,
              password: this.form.password,
              recaptcha: recaptchatoken
            })
            .then(token => {
              this.$store.dispatch('auth/setToken', token).then(() => {
                const success = () => {
                  this.$bvToast.toast('logged in', {
                    variant: 'success',
                    title: 'Success'
                  })
                  this.$router.push({
                    path: this.redirect_uri ? this.redirect_uri : '/dashboard'
                  })
                }
                if (this.$store.state.project.projectId) {
                  this.$store.dispatch('project/getProjectName').then(res => {
                    success()
                  }).catch(err => {
                    console.log(err)
                    this.$bvToast.toast('cannot find current project', {
                      variant: 'danger',
                      title: 'Error'
                    })
                    success()
                  })
                } else {
                  success()
                }
              })
              .catch(err => {
                this.$bvToast.toast(err, {
                  variant: 'danger',
                  title: 'Error'
                })
              })
            }).catch(err => {
              this.$bvToast.toast(err, {
                variant: 'danger',
                title: 'Error'
              })
            })
        })
        .catch(err => {
          this.$bvToast.toast(`got error with recaptcha ${err}`, {
            variant: 'danger',
            title: 'Error'
          })
        })
    }
  }
})
</script>

<style lang="scss"></style>
