import cloneDeepWith from 'lodash.clonedeepwith'

export const clone = (obj) => {
  return cloneDeepWith(obj, (curr) => {
    if (curr && typeof curr === 'object') {
      delete curr.__typename
    }
  })
}
export const arrayMove = (arr, from, to) => {
  if (to >= arr.length) {
    let k = to - arr.length + 1
    while (k--) {
      arr.push(undefined)
    }
  }
  arr.splice(to, 0, arr.splice(from, 1)[0])
}
export const checkDefined = (field) =>
  typeof field !== 'undefined' && field !== null
