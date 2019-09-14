<template>
  <div id="create">
    <b-card no-body>
      <b-card-body>
        <b-form>
          <b-input-group>
            <b-container>
              <b-row>
                <b-col sm>
                  <b-form-input
                    id="name"
                    v-model="name"
                    size="lg"
                    type="text"
                    placeholder="Title"
                  ></b-form-input>
                </b-col>
              </b-row>
              <b-row>
                <b-col sm>
                  <b-form-input
                    id="description"
                    v-model="description"
                    size="sm"
                    type="text"
                    placeholder="Description"
                  ></b-form-input>
                </b-col>
                <b-col sm>
                  <b-form-checkbox
                    v-model="multiple"
                    class="pull-right"
                    name="allow-multiple"
                    switch
                  >
                    Allow Multiple Submissions
                  </b-form-checkbox>
                </b-col>
              </b-row>
            </b-container>
          </b-input-group>
          <hr />
          <draggable v-model="questions" group="questions">
            <div
              v-for="(question, index) in questions"
              :key="`question-${index}`"
              :class="{ 'question-focus': focusIndex === index }"
            >
              <div class="drag-area">
                <no-ssr>
                  <font-awesome-icon class="icon-grip" icon="grip-horizontal" />
                </no-ssr>
              </div>
              <div
                :id="`question-${index}-select-area`"
                @click="evt => focusItem(evt, index)"
              >
                <b-input-group :id="`question-${index}-name-type-input`">
                  <b-container>
                    <b-row>
                      <b-col sm>
                        <b-form-input
                          :id="`question-${index}-name`"
                          v-model="question.name"
                          size="md"
                          type="text"
                          placeholder="Name"
                        ></b-form-input>
                      </b-col>
                      <b-col sm>
                        <b-dropdown
                          v-if="focusIndex === index"
                          :id="`question-type-${index}`"
                          text="type"
                        >
                          <b-dropdown-item-button
                            v-for="(type, indexType) in questionTypes"
                            :key="`question-${index}-select-${indexType}`"
                            @click="evt => selectQuestionType(evt, index, type)"
                            >{{ type.label }}</b-dropdown-item-button
                          >
                        </b-dropdown>
                      </b-col>
                    </b-row>
                  </b-container>
                </b-input-group>
                <b-input-group :id="`question-${index}-content`">
                  <b-container>
                    <div
                      v-if="
                        question.type === questionTypes[0].id ||
                          question.type === questionTypes[1].id
                      "
                    >
                      <b-row
                        v-for="(option, optionIndex) in question.options"
                        :key="`question-${index}-option-${optionIndex}`"
                      >
                        <b-col sm>
                          <b-form-radio
                            v-if="question.type === questionTypes[0].id"
                            disabled
                          >
                            <b-form-input
                              v-model="question.options[optionIndex]"
                              size="sm"
                              type="text"
                              :placeholder="`option ${optionIndex}`"
                            ></b-form-input>
                          </b-form-radio>
                          <b-form-checkbox
                            v-else-if="question.type === questionTypes[1].id"
                            disabled
                          >
                            <b-form-input
                              v-model="question.options[optionIndex]"
                              size="sm"
                              type="text"
                              :placeholder="`option ${optionIndex}`"
                            ></b-form-input>
                          </b-form-checkbox>
                        </b-col>
                        <b-col>
                          <button
                            class="button-link"
                            :disabled="question.options.length <= 1"
                            @click="
                              evt => removeOption(evt, index, optionIndex)
                            "
                          >
                            <no-ssr>
                              <font-awesome-icon class="mr-2" icon="times" />
                            </no-ssr>
                          </button>
                        </b-col>
                      </b-row>
                      <b-form-radio
                        v-if="question.type === questionTypes[0].id"
                        disabled
                      >
                        <button
                          class="button-link"
                          :disabled="
                            question.options[question.options.length - 1]
                              .length === 0
                          "
                          @click="evt => addOption(evt, index)"
                        >
                          Add Radio Option
                        </button>
                      </b-form-radio>
                      <b-form-checkbox
                        v-else-if="question.type === questionTypes[1].id"
                        disabled
                      >
                        <button
                          class="button-link"
                          :disabled="
                            question.options[question.options.length - 1]
                              .length === 0
                          "
                          @click="evt => addOption(evt, index)"
                        >
                          Add Checkbox Option
                        </button>
                      </b-form-checkbox>
                    </div>
                    <b-form-input
                      v-else-if="question.type === questionTypes[2].id"
                      :id="`question-${index}-shortAnswer`"
                      size="sm"
                      type="text"
                      disabled
                      placeholder="short answer"
                    ></b-form-input>
                  </b-container>
                </b-input-group>
              </div>
              <hr />
              <b-input-group>
                <b-container>
                  <b-row>
                    <b-col class="text-right">
                      <button
                        class="button-link"
                        style="display: inline-block;"
                        :disabled="questions.length <= 1"
                        @click="evt => removeQuestion(evt, index)"
                      >
                        <no-ssr>
                          <font-awesome-icon class="mr-2" icon="trash" />
                        </no-ssr>
                      </button>
                      <b-form-checkbox
                        v-model="question.required"
                        style="display: inline-block;"
                        name="required"
                        switch
                      >
                        Required
                      </b-form-checkbox>
                    </b-col>
                  </b-row>
                </b-container>
              </b-input-group>
            </div>
          </draggable>
          <b-button squared variant="primary" @click="addQuestion"
            >Add Question</b-button
          >
          <b-button squared variant="primary" type="submit" @click="submit"
            >Save</b-button
          >
        </b-form>
      </b-card-body>
    </b-card>
  </div>
</template>

<script lang="ts">
import Vue from 'vue'
import clonedeep from 'lodash.clonedeep'

const questionTypes = [
  {
    id: 'radio',
    label: 'Multiple Choice'
  },
  {
    id: 'checkbox',
    label: 'Checkbox'
  },
  {
    id: 'short',
    label: 'Short Answer'
  }
]
const defaultQuestion = {
  name: '',
  type: questionTypes[0].id,
  options: [''],
  required: false
}
export default Vue.extend({
  name: 'Create',
  data() {
    return {
      questionTypes: questionTypes,
      name: '',
      description: '',
      questions: [clonedeep(defaultQuestion)],
      focusIndex: 0,
      multiple: false
    }
  },
  methods: {
    submit(evt) {
      evt.preventDefault()
      /* eslint-disable */
      console.log('submit')
      /* eslint-enable */
    },
    focusItem(evt, questionIndex) {
      evt.preventDefault()
      this.focusIndex = questionIndex
    },
    selectQuestionType(evt, questionIndex, type) {
      evt.preventDefault()
      this.questions[questionIndex].options = ['']
      this.questions[questionIndex].type = type.id
    },
    addQuestion(evt) {
      evt.preventDefault()
      this.questions.push(clonedeep(defaultQuestion))
    },
    removeQuestion(evt, questionIndex) {
      evt.preventDefault()
      this.questions.splice(questionIndex, 1)
    },
    addOption(evt, questionIndex) {
      evt.preventDefault()
      this.questions[questionIndex].options.push('')
    },
    removeOption(evt, questionIndex, optionIndex) {
      evt.preventDefault()
      this.questions[questionIndex].options.splice(optionIndex, 1)
    }
  }
})
</script>

<style lang="scss">
.drag-area {
  padding: 5px;
  cursor: move;
  .icon-grip {
    color: lightgray;
    font-size: 20px;
    margin-left: 37px;
  }
}
</style>
