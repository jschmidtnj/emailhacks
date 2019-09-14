import Vue from 'vue'
import { library } from '@fortawesome/fontawesome-svg-core'
import {
  faGripHorizontal,
  faTimes,
  faTrash
} from '@fortawesome/free-solid-svg-icons'

import { } from '@fortawesome/free-brands-svg-icons'

import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

library.add(
  faGripHorizontal,
  faTimes,
  faTrash
)

Vue.component('font-awesome-icon', FontAwesomeIcon)
