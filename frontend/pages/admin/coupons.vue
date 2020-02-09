<template>
  <div class="container-fluid">
    <div id="admin-cards" class="container">
      <div class="row my-4">
        <div class="col-lg-6 my-2">
          <section class="card h-100 py-0">
            <div class="card-body">
              <b-form @submit="managecoupons" @reset="resetcoupon">
                <span class="card-text">
                  <h2 class="mb-4">{{ mode }} Coupon</h2>
                  <b-form-group>
                    <label class="form-required">Secret</label>
                    <b-input-group>
                      <b-form-input
                        id="secret"
                        v-model="coupon.secret"
                        :state="!$v.coupon.secret.$invalid"
                        type="text"
                        class="form-control"
                        aria-describedby="couponfeedback"
                        placeholder="secret"
                      />
                    </b-input-group>
                    <b-form-invalid-feedback
                      id="couponfeedback"
                      :state="!$v.coupon.secret.$invalid"
                    >
                      <div v-if="!$v.coupon.secret.required">
                        secret is required
                      </div>
                      <div v-else-if="!$v.coupon.secret.minLength">
                        secret must have at least
                        {{ $v.coupon.secret.$params.minLength.min }} characters
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-form-group>
                    <label class="form-required">Amount</label>
                    <b-form-group
                      :prepend="coupon.percent ? '' : '$'"
                      :append="coupon.percent ? '%' : '.00'"
                    >
                      <b-form-input
                        id="amount"
                        v-model="coupon.amount"
                        :state="!$v.coupon.amount.$invalid"
                        type="number"
                        class="form-control"
                        aria-describedby="amountfeedback"
                        placeholder="amount"
                      />
                    </b-form-group>
                    <b-form-invalid-feedback
                      id="amountfeedback"
                      :state="!$v.coupon.amount.$invalid"
                    >
                      <div v-if="!$v.coupon.amount.required">
                        max storage is required
                      </div>
                      <div v-else-if="!$v.coupon.amount.integer">
                        max storage must be an integer
                      </div>
                      <div v-else-if="!$v.coupon.amount.minValue">
                        max storage must be greater than
                        {{ $v.coupon.amount.$params.minValue.min }}
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-form-group>
                    <label class="form-required">Use Percent</label>
                    <b-form-group>
                      <b-form-checkbox
                        id="percent"
                        v-model="coupon.percent"
                        :state="true"
                        class="form-control"
                        switch
                      />
                    </b-form-group>
                  </b-form-group>
                  <b-container class="mt-4">
                    <b-row>
                      <b-col>
                        <b-btn
                          :disabled="$v.coupon.$invalid || submitting"
                          variant="primary"
                          type="submit"
                        >
                          <client-only>
                            <font-awesome-icon
                              class="mr-2 arrow-size-edit"
                              icon="angle-double-right"
                            /> </client-only
                          >Submit
                        </b-btn>
                        <b-btn variant="primary" type="reset" class="mr-4">
                          <client-only>
                            <font-awesome-icon
                              class="mr-2 arrow-size-edit"
                              icon="times"
                            /> </client-only
                          >Clear
                        </b-btn>
                      </b-col>
                    </b-row>
                  </b-container>
                </span>
              </b-form>
            </div>
          </section>
        </div>
        <div class="col-lg-6 my-2">
          <section class="card h-100 py-0">
            <div class="card-body">
              <b-form @submit="searchcoupons" @reset="clearsearch">
                <span class="card-text">
                  <h2 class="mb-4">Search</h2>
                  <b-form-group>
                    <label class="form-required">Query</label>
                    <span>
                      <b-form-input
                        v-model="search"
                        :state="!$v.search.$invalid"
                        type="text"
                        class="form-control mb-2"
                        aria-describedby="searchfeedback"
                        placeholder="search..."
                      />
                    </span>
                    <b-form-invalid-feedback
                      id="searchfeedback"
                      :state="!$v.search.$invalid"
                    >
                      <div v-if="!$v.search.required">query is required</div>
                      <div v-else-if="!$v.search.minLength">
                        query must have at least
                        {{ $v.search.$params.minLength.min }} characters
                      </div>
                    </b-form-invalid-feedback>
                  </b-form-group>
                  <b-btn
                    :disabled="$v.search.$invalid"
                    variant="primary"
                    type="submit"
                    class="mr-4"
                  >
                    <client-only>
                      <font-awesome-icon
                        class="mr-2 arrow-size-edit"
                        icon="angle-double-right"
                      /> </client-only
                    >Search
                  </b-btn>
                  <b-btn variant="primary" type="reset" class="mr-4">
                    <client-only>
                      <font-awesome-icon
                        class="mr-2 arrow-size-edit"
                        icon="times"
                      /> </client-only
                    >Clear
                  </b-btn>
                  <br />
                  <br />
                </span>
              </b-form>
              <b-table
                :items="searchresults"
                :fields="fields"
                show-empty
                stacked="md"
              >
                <template v-slot:cell(secret)="data">
                  {{ data.value }}
                </template>
                <template v-slot:cell(created)="data">
                  {{ formatDate(data.value, 'M/d/yyyy') }}
                </template>
                <template v-slot:cell(id)="data">
                  {{ data.value }}
                </template>
                <template v-slot:cell(actions)="data">
                  <b-button
                    @click="editCoupon(data.item)"
                    size="sm"
                    class="mr-1"
                  >
                    Edit
                  </b-button>
                  <b-button @click="deleteCoupon(data.item)" size="sm">
                    Del
                  </b-button>
                </template>
              </b-table>
            </div>
          </section>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="js">
import Vue from 'vue'
import { validationMixin } from 'vuelidate'
import { required, minLength, minValue, integer } from 'vuelidate/lib/validators'
import { clone } from '~/assets/utils'
// @ts-ignore
const seo = JSON.parse(process.env.seoconfig)
const validAmount = (amount, vm) => amount && amount >= 0 && (!vm.coupon.percent || amount <= 100)
/**
 * coupons edit
 */
const modetypes = {
  add: 'Add',
  edit: 'Edit',
  delete: 'Delete'
}
const defaultCoupon = {
  secret: '',
  amount: 0,
  percent: false
}
export default Vue.extend({
  name: 'Coupons',
  // @ts-ignore
  layout: 'admin',
  mixins: [validationMixin],
  // @ts-ignore
  data() {
    return {
      submitting: false,
      modetypes,
      mode: modetypes.add,
      couponid: null,
      search: '',
      type: 'coupon',
      searchresults: [],
      fields: [
        {
          key: 'secret',
          label: 'Secret',
          sortable: true
        },
        {
          key: 'created',
          label: 'Created',
          sortable: true
        },
        {
          key: 'id',
          label: 'ID',
          sortable: true
        },
        {
          key: 'actions',
          label: 'Actions',
          sortable: false
        }
      ],
      coupon: clone(defaultCoupon)
    }
  },
  // @ts-ignore
  validations: {
    search: {
      required,
      minLength: minLength(3)
    },
    coupon: {
      secret: {
        required,
        minLength: minLength(3)
      },
      amount: {
        required,
        integer,
        minValue: minValue(0),
        validAmount
      },
      percent: {}
    }
  },
  // @ts-ignore
  head() {
    const title = 'Admin Edit Coupon'
    const description = 'admin page for editing coupons'
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
  /* eslint-disable */
  methods: {
    mongoidToDate(id) {
      return parseInt(id.substring(0, 8), 16) * 1000
    },
    editCoupon(searchresult) {
      this.couponid = searchresult.id
      // get coupon data first
      this.$apollo.query({
        query: gql`
          query coupon($id: String!) {
            coupon(id: $id) {
              secret,
              amount,
              percent
            }
          }`,
          variables: {id: this.couponid},
          fetchPolicy: 'network-only'
        }).then(({ data }) => {
          this.mode = this.modetypes.edit
          this.coupon = data.coupon
          this.$bvToast.toast(`edit coupon with id ${this.couponid}`, {
            variant: 'success',
            title: 'Success'
          })
        }).catch(err => {
          console.error(err)
          this.$bvToast.toast(`found error: ${err.message}`, {
            variant: 'danger',
            title: 'Error'
          })
        })
    },
    deleteCoupon(searchresult) {
      const id = searchresult.id
      this.$apollo.mutate({mutation: gql`
        mutation deleteCoupon($id: String!){deleteCoupon(id: $id){id} }
        `, variables: {id: id}})
        .then(({ data }) => {
          this.searchresults.splice(
            this.searchresults.indexOf(searchresult),
            1
          )
          this.$bvToast.toast('coupon deleted', {
            variant: 'success',
            title: 'Success'
          })
        }).catch(err => {
          console.error(err)
          this.$bvToast.toast(`found error: ${err.message}`, {
            variant: 'danger',
            title: 'Error'
          })
        })
    },
    searchcoupons(evt) {
      evt.preventDefault()
      this.$apollo.query({
        query: gql`
          query coupons() {
            coupons() {
              secret,
              id
            }
          }`,
          variables: {},
          fetchPolicy: 'network-only'
        }).then(({ data }) => {
          const coupons = data.coupons
          coupons.map(
            coupon => {
              coupon.created = this.mongoidToDate(coupon.id)
            }
          )
          this.searchresults = coupons
          this.$bvToast.toast(`found ${this.searchresults.length} result${
              this.searchresults.length === 1 ? '' : 's'
            }`, {
            variant: 'success',
            title: 'Success'
          })
        }).catch(err => {
          console.error(err)
          this.$bvToast.toast(`found error: ${err.message}`, {
            variant: 'danger',
            title: 'Error'
          })
        })
    },
    clearsearch(evt) {
      if (evt) evt.preventDefault()
      this.search = ''
      this.searchresults = []
    },
    resetcoupon(evt) {
      if (evt) evt.preventDefault()
      this.coupon = clone(defaultCoupon)
      this.mode = this.modetypes.add
      this.couponid = null
    },
    formatDate(dateUTC) {
      return formatRelative(dateUTC, new Date())
    },
    managecoupons(evt) {
      evt.preventDefault()
      let couponid = this.couponid
      this.submitting = true
      const onSuccess = () => {
        this.$bvToast.toast(`${this.mode}ed coupon with id ${couponid}`, {
          variant: 'danger',
          title: 'Error'
        })
        this.submitting = false
        this.resetcoupon(null)
      }
      if (this.mode === this.modetypes.add) {
        this.$apollo.mutate({mutation: gql`
          mutation addCoupon($secret: String!, $amount: Int!, $percent: Boolean!) {
            addCoupon(secret: $secret, amount: $amount, percent: $percent) {
              id
            }
          }`, variables: {secret: this.coupon.secret, amount: this.coupon.amount, percent: this.coupon.percent}})
          .then(({ data }) => {
            couponid = data.addCoupon.id
            onSuccess()
          }).catch(err => {
            console.error(err)
            this.$bvToast.toast(`found error: ${err.message}`, {
              variant: 'danger',
              title: 'Error'
            })
          })
      } else {
        this.$apollo.mutate({mutation: gql`
          mutation updateCoupon($id: String!, $secret: String!, $amount: Int!, $percent: Boolean!)
          {updateCoupon(id: $id, secret: $secret, amount: $amount, percent: $percent){id} }
          `, variables: {id: this.couponid, secret: this.coupon.secret, amount: this.coupon.amount, percent: this.coupon.percent}})
          .then(({ data }) => {
            onSuccess()
          }).catch(err => {
            console.error(err)
            this.$bvToast.toast(`found error: ${err.message}`, {
              variant: 'danger',
              title: 'Error'
            })
          })
      }
    }
  }
})
</script>

<style lang="scss"></style>

