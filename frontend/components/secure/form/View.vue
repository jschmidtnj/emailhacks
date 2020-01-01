<template>
  <div id="create">
    <div v-if="!loading">
      <b-card no-body class="card-data shadow-lg pb-4">
        <b-card-body>
          <b-form @submit.prevent>
            <h3 class="mb-4">
              {{ name }}
            </h3>
            <div
              v-for="(item, index) in items"
              :key="`item-${index}`"
              :class="{ 'item-focus': focusIndex === index }"
            >
              <hr class="separate-items" />
              <span
                :id="`item-${index}-select-area`"
                v-touch:start="(evt) => focusItem(evt, index)"
              >
                <b-container v-if="item.type !== itemTypes[3]">
                  <h4>{{ item.question }}</h4>
                </b-container>
                <b-container v-else>
                  <p v-html="item.text" />
                </b-container>
                <b-input-group
                  v-if="
                    item.type !== itemTypes[3] && item.type !== itemTypes[6]
                  "
                  :id="`item-${index}-question`"
                >
                  <b-container>
                    <div
                      v-if="
                        item.type === itemTypes[0] || item.type === itemTypes[1]
                      "
                    >
                      <b-row
                        v-for="(option, optionIndex) in item.options"
                        :key="`item-${index}-option-${optionIndex}`"
                        class="mt-2 mb-2"
                        style="max-width:30rem;"
                      >
                        <b-col>
                          <b-form-radio v-if="item.type === itemTypes[0]">{{
                            item.options[optionIndex]
                          }}</b-form-radio>
                          <b-form-checkbox
                            v-else-if="item.type === itemTypes[1]"
                            >{{ item.options[optionIndex] }}</b-form-checkbox
                          >
                        </b-col>
                      </b-row>
                    </div>
                    <b-form-textarea
                      v-else-if="item.type === itemTypes[2]"
                      :id="`item-${index}-shortAnswer`"
                      class="mt-2 mb-2"
                      rows="3"
                      max-rows="8"
                      style="max-width:30rem;"
                    />
                    <div
                      v-else-if="item.type === itemTypes[4]"
                      class="mt-2 mb-2"
                    >
                      <b-form-checkbox
                        :id="`item-${index}-red-green`"
                        style="display: inline-block;"
                        name="red-green"
                        switch
                      />
                    </div>
                    <div
                      v-else-if="item.type === itemTypes[5]"
                      class="mt-2 mb-2"
                    >
                      <b-form-file
                        placeholder="Choose a file or drop it here..."
                        drop-placeholder="Drop file here..."
                        style="max-width:30rem;"
                      />
                    </div>
                    <div
                      v-else-if="item.type === itemTypes[6]"
                      class="mt-2 mb-2"
                    >
                      <b-container>
                        <b-row>
                          <b-col>
                            <a
                              v-if="items[index].file"
                              :href="getFileURL(index)"
                              :download="items[index].file.name"
                              class="mt-2 mb-2"
                              >Download</a
                            >
                          </b-col>
                        </b-row>
                      </b-container>
                    </div>
                  </b-container>
                </b-input-group>
              </span>
            </div>
          </b-form>
        </b-card-body>
      </b-card>
      <b-container style="margin-top: 3rem; margin-bottom: 2rem;">
        <b-row>
          <b-col class="text-right">
            <b-button
              @click="submit"
              pill
              variant="primary"
              class="submit-button shadow-lg"
            >
              <client-only>
                <font-awesome-icon size="3x" icon="paper-plane" />
              </client-only>
            </b-button>
          </b-col>
        </b-row>
      </b-container>
    </div>
    <page-loading v-else :loading="true" />
  </div>
</template>

<script lang="js">
import Vue from 'vue'
import gql from 'graphql-tag'
import PageLoading from '~/components/PageLoading.vue'
import { noneAccessType } from '~/assets/config'
const itemTypes = ['radio', 'checkbox', 'short', 'text',
  'redgreen', 'fileupload', 'fileattachment']
export default Vue.extend({
  name: 'ViewForm',
  components: {
    PageLoading
  },
  props: {
    formId: {
      type: String,
      default: null
    },
    projectId: {
      type: String,
      default: null
    }
  },
  data() {
    return {
      name: '',
      items: [],
      multiple: false,
      files: [],
      focusIndex: 0,
      itemTypes,
      loading: true,
      isPublic: false
    }
  },
  mounted() {
    if (this.formId) {
      this.$apollo.query({query: gql`
        query form($id: String!){form(id: $id){name,items{question,type,options,text,required,files},multiple,files{id,name,width,height,type}} }
        `, variables: {id: this.formId}})
        .then(({ data }) => {
          console.log(data.form)
          this.name = data.form.name
          this.isPublic = data.form.public === noneAccessType
          this.$store.commit('auth/setRedirectLogin', this.isPublic)
          this.items = data.form.items
          this.multiple = data.form.multiple
          this.files = data.form.files
          this.loading = false
        }).catch(err => {
          console.log(err.message)
          this.$toasted.global.error({
            message: `found error: ${err.message}`
          })
        })
    }
  },
  methods: {
    getFileURL(itemIndex) {
      return ''
    },
    focusItem(evt, itemIndex) {
      this.focusIndex = itemIndex
    },
    submit(evt) {
      evt.preventDefault()
      console.log('submitted!')
    }
  }
})
</script>

<style lang="scss">
.separate-items {
  margin-top: 2rem;
  margin-bottom: 2rem;
}
.submit-button {
  height: 6rem;
  width: 6rem;
  text-align: center;
  line-height: 50%;
}
.card-data {
  @media (min-width: 600px) {
    max-width: 50rem;
  }
}
</style>
