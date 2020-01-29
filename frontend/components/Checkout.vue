<template>
  <b-card>
    <b-container>
      <b-row>
        <b-col>
          <b-form @submit="pay">
            <b-row>
              <b-form-group class="col-lg-4">
                <label class="form-required">First Name</label>
                <span>
                  <b-form-input
                    id="billfirstname"
                    v-model="billing.firstname"
                    :state="!$v.billing.firstname.$invalid"
                    type="text"
                    class="form-control"
                    aria-describedby="billfirstnamefeedback"
                    placeholder
                  />
                </span>
                <b-form-invalid-feedback
                  id="billfirstnamefeedback"
                  :state="!$v.billing.firstname.$invalid"
                >
                  <div v-if="!$v.billing.firstname.required">
                    first name is required
                  </div>
                  <div v-else-if="!$v.billing.firstname.minLength">
                    first name must have at least
                    {{ $v.billing.firstname.$params.minLength.min }} characters
                  </div>
                </b-form-invalid-feedback>
              </b-form-group>
              <b-form-group class="col-lg-4">
                <label class="form-required">Last Name</label>
                <span>
                  <b-form-input
                    id="billlastname"
                    v-model="billing.lastname"
                    :state="!$v.billing.lastname.$invalid"
                    type="text"
                    class="form-control"
                    aria-describedby="billlastnamefeedback"
                    placeholder
                  />
                </span>
                <b-form-invalid-feedback
                  id="billlastnamefeedback"
                  :state="!$v.billing.lastname.$invalid"
                >
                  <div v-if="!$v.billing.lastname.required">
                    last name is required
                  </div>
                  <div v-else-if="!$v.billing.lastname.minLength">
                    last name must have at least
                    {{ $v.billing.lastname.$params.minLength.min }} characters
                  </div>
                </b-form-invalid-feedback>
              </b-form-group>
            </b-row>
            <!-- Name Group -->
            <b-form-group>
              <label class="form">Firm / Company / Employer</label>
              <span>
                <b-form-input
                  id="billcompany"
                  v-model="billing.company"
                  :state="!$v.billing.company.$invalid"
                  type="text"
                  class="form-control"
                  aria-describedby="billcompanyfeedback"
                  placeholder
                />
              </span>
              <b-form-invalid-feedback
                id="billcompanyfeedback"
                :state="!$v.billing.company.$invalid"
              >
                <div v-if="!$v.billing.company.required">
                  company is required
                </div>
                <div v-else-if="!$v.billing.company.minLength">
                  company must have at least
                  {{ $v.billing.company.$params.minLength.min }} characters
                </div>
              </b-form-invalid-feedback>
            </b-form-group>
            <b-row>
              <client-only>
                <multiselect
                  v-model="selectedCountry"
                  :options="countryOptions"
                  :multiple="false"
                  @change="setCountryGetData"
                  label="countryName"
                />
              </client-only>
              <vue-google-autocomplete
                id="map"
                ref="address"
                v-on:placechanged="getAddressData"
                :country="selectedCountry.countryCode"
                classname="form-control"
                placeholder="Please type your address"
              />
            </b-row>
            <b-row>
              <b-form-group class="col-lg-4">
                <label class="form-required">Phone</label>
                <span>
                  <b-form-input
                    id="billphone"
                    v-model="billing.phone"
                    :state="!$v.billing.phone.$invalid"
                    type="text"
                    class="form-control"
                    aria-describedby="billphonefeedback"
                    placeholder
                  />
                </span>
                <b-form-invalid-feedback
                  id="billphonefeedback"
                  :state="!$v.billing.phone.$invalid"
                >
                  <div v-if="!$v.billing.phone.required">phone is required</div>
                  <div v-else-if="!$v.billing.phone.validPhone">
                    phone is invalid
                  </div>
                </b-form-invalid-feedback>
              </b-form-group>
              <b-form-group class="col-lg-8">
                <label class="form-required">Email Address</label>
                <span>
                  <b-form-input
                    id="billemail"
                    v-model="billing.email"
                    :state="!$v.billing.email.$invalid"
                    type="text"
                    class="form-control"
                    placeholder="email"
                    aria-describedby="billemailfeedback"
                  />
                </span>
                <b-form-invalid-feedback
                  id="billemailfeedback"
                  :state="!$v.billing.email.$invalid"
                >
                  <div v-if="!$v.billing.email.required">email is required</div>
                  <div v-else-if="!$v.billing.email.email">
                    email is invalid
                  </div>
                </b-form-invalid-feedback>
              </b-form-group>
            </b-row>
            <!-- Billing -->
            <h5 class="my-4">Payment</h5>
            <card
              :class="{ payment }"
              :stripe="stripetoken"
              :options="stripeOptions"
              @change="payment = $event.complete"
              class="stripe-card"
            />
            <h6 v-if="coupon">
              Using coupon {{ coupon.secret }} for
              {{
                coupon.percent
                  ? `${coupon.amount}%`
                  : formatCurrency(coupon.amount)
              }}
              off
            </h6>
            <b-row>
              <b-form-group class="col-lg-4">
                <label class="form-required">Discounts and Coupons</label>
                <span>
                  <b-form-input
                    id="couponinput"
                    v-model="potentialCoupon"
                    :state="validCoupon"
                    type="text"
                    class="form-control"
                    aria-describedby="couponfeedback"
                    placeholder="Enter code"
                  />
                </span>
                <b-form-invalid-feedback
                  id="couponfeedback"
                  v-if="potentialCoupon"
                  :state="validCoupon"
                >
                  <div v-if="!validCoupon">coupon is invalid</div>
                </b-form-invalid-feedback>
              </b-form-group>
              <b-btn
                @click="checkCoupon"
                :disabled="!potentialCoupon"
                class="mt-4"
              >
                Apply
              </b-btn>
              <b-btn @click="removeCoupon" v-if="coupon" class="mt-4">
                Remove Coupon
              </b-btn>
            </b-row>
            <b-btn
              @click="pay"
              :disabled="
                !payment ||
                  $v.billing.$invalid ||
                  !$store.state.purchase.plan ||
                  $store.state.purchase.products.length === 0
              "
              class="mt-4"
            >
              Pay {{ total }}
            </b-btn>
          </b-form>
        </b-col>
      </b-row>
    </b-container>
  </b-card>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import Multiselect from 'vue-multiselect'
import { validationMixin } from 'vuelidate'
import { required, email, minLength } from 'vuelidate/lib/validators'
import { Card, createSource } from 'vue-stripe-elements-plus'
import { getCodes, getName } from 'country-list'
import { formatLocaleCurrency } from 'country-currency-map'
import flag from 'country-code-emoji'
import { regex } from '~/assets/config'
const validPhone = val => regex.phone.test(val)
const stripetoken = JSON.parse(process.env.stripeconfig).clienttoken
const countryOptions = getCodes().map((code) => {
  return {
    countryCode: code,
    countryName: `${flag(code)} ${getName(code)}`
  }
})
// log in again after payment is complete for different pass
// check this out: https://alligator.io/vuejs/stripe-elements-vue-integration/
export default Vue.extend({
  name: 'Checkout',
  components: {
    Card,
    Multiselect
  },
  // @ts-ignore
  head() {
    const mapsautoapikey = process.env.mapsautoapikey
    return {
      script: [
        {
          src: `https://maps.googleapis.com/maps/api/js?key=${mapsautoapikey}&libraries=places`
        }
      ]
    }
  },
  mixins: [validationMixin],
  data() {
    return {
      payment: false,
      stripetoken,
      countryOptions,
      selectedCountry: null,
      coupon: null,
      potentialCoupon: '',
      validCoupon: false,
      billing: {
        firstname: '',
        lastname: '',
        company: '',
        address1: '',
        address2: '',
        city: '',
        state: '',
        zip: '',
        phone: '',
        email: ''
      },
      stripeOptions: {
        // see https://stripe.com/docs/stripe.js#element-options for details
        hidePostalCode: true,
        style: {
          base: {
            fontSize: '18px',
            color: '#32325d',
            fontSmoothing: 'antialiased',
            '::placeholder': {
              color: '#ccc'
            }
          },
          invalid: {
            color: '#e5424d',
            ':focus': {
              color: '#303238'
            }
          }
        }
      }
    }
  },
  // @ts-ignore
  validations: {
    billing: {
      firstname: {
        required,
        minLength: minLength(3)
      },
      lastname: {
        required,
        minLength: minLength(3)
      },
      company: {
        required,
        minLength: minLength(3)
      },
      phone: {
        required,
        validPhone
      },
      email: {
        required,
        email
      }
    }
  },
  computed: {
    total() {
      return this.formatCurrency(this.$store.getters.purchase.total * this.$store.state.auth.exchangeRate, this.$store.state.auth.currency)
    }
  },
  mounted() {
    this.$apollo
      .query({
        query: gql`
          query account {
            account {
              billing {
                firstname
                lastname
                company
                address1
                address
                city
                state
                zip
                country
                phone
                email
              }
            }
          }
        `,
        variables: {},
        fetchPolicy: 'network-only'
      })
      .then(({ data }) => {
        if (data.billing) {
          this.billing = data.billing
        } else {
          this.$bvToast.toast('cannot find billing info', {
            variant: 'danger',
            title: 'Error'
          })
        }
      })
      .catch((err) => {
        this.$bvToast.toast(err, {
          variant: 'danger',
          title: 'Error'
        })
      })
    this.$store.dispatch('auth/getCountry').then(() => {
      let defaultCountryIndex = countryOptions.findIndex(
        (country) => country.countryCode === this.$store.state.auth.country
      )
      if (!defaultCountryIndex) {
        defaultCountryIndex = 0
      }
      this.selectedCountry = countryOptions[defaultCountryIndex]
    }).catch(err => {
      this.$bvToast.toast(err, {
        variant: 'danger',
        title: 'Error'
      })
    })
  },
  methods: {
    checkCoupon(evt) {
      evt.preventDefault()
      this.$apollo.query({query: gql`
        query checkCoupon($secret: String!) {
          checkCoupon(secret: $secret) {
            secret
            amount
            percent
          }
        }
        `, variables: {
          secret: this.potentialCoupon
        }})
        .then(({ data }) => {
          if (!data.checkCoupon) {
            this.$bvToast.toast('invalid coupon', {
              variant: 'danger',
              title: 'Error'
            })
          } else {
            this.validCoupon = true
            this.coupon = data.checkCoupon
            this.$bvToast.toast('valid coupon', {
              variant: 'success',
              title: 'Success'
            })
          }
        }).catch(err => {
          this.$bvToast.toast(`found error: ${err.message}`, {
            variant: 'danger',
            title: 'Error'
          })
        })
    },
    removeCoupon(evt) {
      evt.preventDefault()
      this.coupon = null
      this.validCoupon = false
    },
    formatCurrency(amount) {
      return formatLocaleCurrency(amount)
    },
    setCountryGetData() {
      this.$store.dispatch('auth/setCountryGetData', this.selectedCountry).then(() => {
        console.log('updated to selected country')
      }).catch(err => {
        this.$bvToast.toast(err, {
          variant: 'danger',
          title: 'Error'
        })
      })
    },
    getAddressData(addressData, placeResultData, id) {
      // TODO - process data and output to address
      console.log(addressData)
    },
    pay(evt) {
      evt.preventDefault()
      // createToken returns a Promise which resolves in a result object with
      // either a token or an error key.
      // See https://stripe.com/docs/api#tokens for the token object.
      // See https://stripe.com/docs/api#errors for the error object.
      // More general https://stripe.com/docs/stripe.js#stripe-create-token.
      createSource({
        type: 'card',
        currency: this.$store.state.auth.currency,
        owner: {
          address: {
            line1: this.billing.address1,
            line2: this.billing.address2,
            city: this.billing.city,
            state: this.billing.state,
            postal_code: this.billing.zip
          },
          name: `${this.billing.firstname} ${this.billing.lastname}${this.billing.company ? ` - ${this.billing.company}` : ''}`,
          phone: this.billing.phone,
          email: this.billing.email
        }
      })
        .then((data) => {
          if (data.errors) {
            this.$bvToast.toast(`found error(s): ${data.errors}`, {
              variant: 'danger',
              title: 'Error'
            })
          } else if (!data.source && !data.source.id) {
            this.$bvToast.toast('cannot find stripe token', {
              variant: 'danger',
              title: 'Error'
            })
          } else {
            console.log(data.source)
            console.log(data.source.id)
            const cardToken = data.source.id
            let numPurchaseComplete = 0
            const numPurchases = (this.$store.state.purchase.plan ? 1 : 0) + this.$store.state.purchase.products.length
            const success = () => {
              this.$bvToast.toast('completed purchase', {
                variant: 'success',
                title: 'Success'
              })
            }
            const purchase = (product, interval) => {
              this.$apollo.mutate({mutation: gql`
                mutation purchase($product: String!, $interval: String!, $cardToken: String!, $coupon: String) {
                  purchase(product: $product, interval: $interval, cardToken: $cardToken, coupon: $coupon) {
                    id
                  }
                }
                `, variables: {
                  product,
                  interval,
                  cardToken,
                  coupon: this.coupon.amount
                }})
                .then(({ data }) => {
                  numPurchaseComplete++
                  if (numPurchaseComplete === numPurchases) {
                    success()
                  }
                }).catch(err => {
                  this.$bvToast.toast(`found error: ${err.message}`, {
                    variant: 'danger',
                    title: 'Error'
                  })
                })
            }
            if (this.$store.state.purchase.plan) {
              purchase(this.$store.state.purchase.plan.product, this.$store.state.purchase.plan.interval)
            }
            this.$store.state.purchase.products.forEach(item => {
              purchase(item.product, item.interval)
            })
          }
        })
        .catch((err) => {
          this.$bvToast.toast(err, {
            variant: 'danger',
            title: 'Error'
          })
        })
    }
  }
})
</script>

<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>

<style lang="scss"></style>
