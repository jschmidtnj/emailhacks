<template>
  <b-card v-if="!loading">
    <cart />
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
              <country-selector
                @select="(country) => (billing.country = country)"
              />
              <currency-selector
                @select="(currency) => (billing.currency = currency)"
              />
              <client-only>
                <form
                  @submit="(evt) => evt.preventDefault()"
                  autocomplete="disabled"
                >
                  <vue-google-autocomplete
                    id="map"
                    ref="address"
                    v-if="billing.country"
                    :value="initialAddress"
                    v-on:placechanged="getAddressData"
                    :country="billing.country"
                    classname="form-control"
                    placeholder="Please type your address"
                  />
                </form>
              </client-only>
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
            <div id="stripe-card-element">
              <!-- A Stripe Element will be inserted here. -->
            </div>
            <h6 v-if="coupon && billing.country && billing.currency">
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
                $v.billing.$invalid ||
                  !selectedAddress ||
                  (!$store.state.purchase.plan &&
                    $store.state.purchase.products.length === 0)
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
import VueGoogleAutocomplete from 'vue-google-autocomplete'
import { validationMixin } from 'vuelidate'
import { required, email, minLength } from 'vuelidate/lib/validators'
import { formatLocaleCurrency } from 'country-currency-map'
import { getCode } from 'country-list'
import Cart from '~/components/Cart.vue'
import CountrySelector from '~/components/CountrySelector.vue'
import CurrencySelector from '~/components/CurrencySelector.vue'
import { regex } from '~/assets/config'
const validPhone = val => regex.phone.test(val)
const stripetoken = JSON.parse(process.env.stripeconfig).clienttoken
// log in again after payment is complete for different pass
// check this out: https://alligator.io/vuejs/stripe-elements-vue-integration/
export default Vue.extend({
  name: 'Checkout',
  components: {
    VueGoogleAutocomplete,
    Cart,
    CountrySelector,
    CurrencySelector
  },
  mixins: [validationMixin],
  data() {
    return {
      loading: true,
      stripe: null,
      stripeCard: null,
      selectedAddress: false,
      coupon: null,
      potentialCoupon: '',
      validCoupon: false,
      billing: {
        firstname: '',
        lastname: '',
        company: '',
        address1: '',
        city: '',
        state: '',
        country: '',
        zip: '',
        phone: '',
        email: '',
        currency: ''
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
      return this.formatCurrency(this.$store.getters['purchase/total'] * this.$store.state.auth.exchangeRate)
    }
  },
  mounted() {
    let gotBilling = false
    let gotMapsScript = false
    let gotGoogleAPIScript = false
    const onFinishedLoading = () => {
      if (gotBilling && gotMapsScript && gotGoogleAPIScript) {
        if (this.billing.address1 && this.billing.address1.length > 0) {
          this.initialAddress = `${this.billing.address1}, ${this.billing.city}, ${this.billing.state} ${this.billing.zip}`
          this.selectedAddress = true
        } else {
          this.initialAddress = null
        }
        this.loading = false
        this.$nextTick(() => {
          // eslint-disable-next-line
          this.stripe = Stripe(stripetoken)
          this.stripeCard = this.stripe.elements().create('card', {
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
          })
          this.stripeCard.mount('#stripe-card-element')
        })
      }
    }
    if (!window.google) {
      const apiScript = document.createElement('script')
      apiScript.onload = () => {
        gotGoogleAPIScript = true
        onFinishedLoading()
      }
      apiScript.type = 'text/javascript'
      apiScript.src = 'https://apis.google.com/js/api.js'
      document.head.appendChild(apiScript)
      const mapsautoapikey = process.env.mapsautoapikey
      const mapsScript = document.createElement('script')
      mapsScript.onload = () => {
        gotMapsScript = true
        onFinishedLoading()
      }
      mapsScript.type = 'text/javascript'
      mapsScript.src = `https://maps.googleapis.com/maps/api/js?key=${mapsautoapikey}&libraries=places`
      document.head.appendChild(mapsScript)
    } else {
      gotGoogleAPIScript = true
      gotMapsScript = true
    }
    this.$apollo
      .query({
        query: gql`
          query account {
            account {
              id
              billing {
                firstname
                lastname
                company
                address1
                city
                state
                zip
                phone
                email
                country
                currency
              }
            }
          }
        `,
        variables: {},
        fetchPolicy: 'network-only'
      })
      .then(({ data }) => {
        if (data.account && data.account.billing) {
          this.billing = data.account.billing
          gotBilling = true
          onFinishedLoading()
        } else {
          this.$bvToast.toast('cannot find billing info', {
            variant: 'danger',
            title: 'Error'
          })
        }
      })
      .catch((err) => {
        console.error(err)
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
      return formatLocaleCurrency(amount, this.$store.state.auth.currency)
    },
    getAddressData(addressData, placeResultData, id) {
      this.billing.address1 = `${addressData.street_number} ${addressData.route}`
      this.billing.city = addressData.locality
      this.billing.state = addressData.administrative_area_level_1
      this.billing.zip = addressData.postal_code
      this.billing.country = addressData.country
      console.log(addressData)
      this.selectedAddress = true
    },
    pay(evt) {
      evt.preventDefault()
      this.$apollo.mutate({mutation: gql`
        mutation changeBilling($firstname: String, $lastname: String, $company: String, $address1: String, $city: String, $state: String, $zip: String, $country: String, $phone: String, $email: String, $currency: String) {
          changeBilling(firstname: $firstname, lastname: $lastname, company: $company, address1: $address1, city: $city, state: $state, zip: $zip, country: $country, phone: $phone, email: $email, currency: $currency) {
            id
          }
        }
        `, variables: {
          firstname: this.billing.firstname,
          lastname: this.billing.lastname,
          company: this.billing.company,
          address1: this.billing.address1,
          city: this.billing.city,
          state: this.billing.state,
          zip: this.billing.zip,
          country: this.billing.country,
          phone: this.billing.phone,
          email: this.billing.email,
          currency: this.billing.currency,
        }})
      .then(({ data }) => {
        const success = () => {
          this.$bvToast.toast('completed purchase', {
            variant: 'success',
            title: 'Success'
          })
        }
        const purchase = ( { productIndex, planIndex }, cardToken) => {
          return new Promise((resolve, reject) => {
            this.$apollo.mutate({mutation: gql`
             mutation purchase($product: String!, $interval: String!, $cardToken: String!, $coupon: String) {
                purchase(product: $product, interval: $interval, cardToken: $cardToken, coupon: $coupon)
              }
              `, variables: {
                product: this.$store.state.purchase.productOptions[productIndex].id,
                interval: this.$store.state.purchase.productOptions[productIndex].plans[planIndex].interval,
                cardToken,
                coupon: this.coupon ? this.coupon.amount : null
              }})
              .then(({ data }) => {
                resolve(data)
              }).catch(err => {
                reject(err)
              })
          })
        }
        let numPurchaseComplete = 0
        const numPurchases = (this.$store.state.purchase.plan ? 1 : 0) + this.$store.state.purchase.products.length
        const paymentParams = {
          type: 'card',
          card: this.stripeCard,
          billing_details: {
            address: {
              line1: this.billing.address1,
              city: this.billing.city,
              state: this.billing.state,
              postal_code: this.billing.zip,
              country: getCode(this.billing.country)
            },
            name: `${this.billing.firstname} ${this.billing.lastname}${this.billing.company ? ` - ${this.billing.company}` : ''}`,
            phone: this.billing.phone,
            email: this.billing.email
          }
        }
        if (this.$store.state.purchase.plan) {
          this.stripe.createPaymentMethod(paymentParams)
            .then((data) => {
              if (data.error) {
                console.error(data.error)
                this.$bvToast.toast(`found error(s): ${data.error.message}`, {
                  variant: 'danger',
                  title: 'Error'
                })
              } else if (!(data.paymentMethod && data.paymentMethod.id)) {
                console.log(data)
                this.$bvToast.toast('cannot find stripe token', {
                  variant: 'danger',
                  title: 'Error'
                })
              } else {
                const cardtoken = data.paymentMethod.id
                purchase(this.$store.state.purchase.plan, cardtoken).then(res => {
                  numPurchaseComplete++
                  if (numPurchaseComplete === numPurchases) {
                    success()
                  }
                }).catch(err => {
                  this.$bvToast.toast(err.message, {
                    variant: 'danger',
                    title: 'Error'
                  })
                })
              }
            })
            .catch((err) => {
              console.error(err)
              this.$bvToast.toast(err, {
                variant: 'danger',
                title: 'Error'
              })
            })
        }
        this.$store.state.purchase.products.forEach(item => {
          purchase(item, null).then(res => {
            if (res.purchase) {
              this.stripe.confirmCardPayment(
                res.purchase,
                {
                  payment_method: paymentParams
                }
              ).then((res) => {
                if (res.error) {
                  this.$bvToast.toast(res.error, {
                    variant: 'danger',
                    title: 'Error'
                  })
                } else {
                  numPurchaseComplete++
                  if (numPurchaseComplete === numPurchases) {
                    success()
                  }
                }
              }).catch(err => {
                this.$bvToast.toast(err.message, {
                  variant: 'danger',
                  title: 'Error'
                })
              })
            } else {
              this.$bvToast.toast('cannot find purchase key', {
                variant: 'danger',
                title: 'Error'
              })
            }
          }).catch(err => {
            this.$bvToast.toast(err.message, {
              variant: 'danger',
              title: 'Error'
            })
          })
        })
      }).catch(err => {
        this.$bvToast.toast(err.message, {
          variant: 'danger',
          title: 'Error'
        })
      })
    }
  }
})
</script>

<style lang="scss"></style>
