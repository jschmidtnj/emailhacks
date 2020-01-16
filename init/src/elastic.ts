import elasticsearch = require('elasticsearch')
import { elasticuri } from './config'

/**
 * elastic functions - initialize elasticsearch
 */

export const blogMappings = {
  properties: {
    title: {
      type: 'keyword'
    },
    caption: {
      type: 'keyword'
    },
    author: {
      type: 'keyword'
    },
    color: {
      type: 'text'
    },
    tags: {
      type: 'keyword'
    },
    categories: {
      type: 'keyword'
    },
    content: {
      type: 'text'
    },
    views: {
      type: 'integer'
    },
    created: {
      type: 'date',
      format: 'epoch_millis'
    },
    updated: {
      type: 'date',
      format: 'epoch_millis'
    },
    heroimage: {
      type: 'nested'
    },
    tileimage: {
      type: 'nested'
    },
    files: {
      type: 'nested'
    },
    comments: {
      type: 'nested'
    }
  }
}

export const formMappings = {
  properties: {
    name: {
      type: 'text'
    },
    items: {
      type: 'nested'
    },
    multiple: {
      type: 'boolean'
    },
    access: {
      type: 'object'
    },
    views: {
      type: 'integer'
    },
    responses: {
      type: 'integer'
    },
    public: {
      type: 'keyword'
    },
    files: {
      type: 'nested'
    },
    created: {
      type: 'date',
      format: 'epoch_millis'
    },
    updated: {
      type: 'date',
      format: 'epoch_millis'
    },
    project: {
      type: 'keyword'
    }
  }
}

export const projectMappings = {
  properties: {
    name: {
      type: 'text'
    },
    forms: {
      type: 'integer'
    },
    access: {
      type: 'object'
    },
    views: {
      type: 'integer'
    },
    public: {
      type: 'keyword'
    },
    created: {
      type: 'date',
      format: 'epoch_millis'
    },
    updated: {
      type: 'date',
      format: 'epoch_millis'
    }
  }
}

const addRemoveAccessScript = `
for (int i = 0; i < access.length; i++) {
  bool cont = true;
  if (access[i].type != null) {
    if (access[i].type == 'none') {
      if (ctx._source.access[access[i].id] != null) {
        ctx._source.access[access[i].id].remove(access[i].id);
      }
      cont = false;
    } else {
      if (ctx._source.access[access[i].id] != null) {
        ctx._source.access[access[i].id].type = access[i].type;
      } else {
        ctx._source.access[access[i].id] = {
          'type': access[i].type
        }
      }
    }
  }
  if (cont) {
    if (access[userIDString].categories != null) {
      ctx._source.access[userIDString].categories = categories
    }
    if (access[userIDString].tags != null) {
      ctx._source.access[userIDString].tags = tags
    }
  }
}
`.replace('\n', '')

const addRemoveAccessScriptName = 'addRemoveAccess'

const indexsettings = {
  number_of_shards: 1,
  number_of_replicas: 0
}

const writeclient = new elasticsearch.Client({
  host: elasticuri
})

export const initializeElasticScripts = () => {
  return new Promise((resolve, reject) => {
    writeclient
      .deleteScript({
        id: addRemoveAccessScriptName
      }).then(res1 => {
        console.log(`deleted script res: ${res1}`)
      }).catch(err => {
        console.log(`got error: ${err}`)
      }).then(() => {
        writeclient.putScript({
          id: addRemoveAccessScriptName,
          body: addRemoveAccessScript
        }).then(res => {
          resolve(res)
        }).catch(err => {
          reject(err)
        })
      })
  })
}

export const initializeElasticMappings = (indexname, doctype, mappings) => {
  return new Promise((resolve, reject) => {
    writeclient
      .ping({
        requestTimeout: 1000
      })
      .then(() => {
        console.log(`able to ping writeclient`)
        writeclient.indices
          .putSettings({
            index: indexname,
            body: {
              index: indexsettings
            }
          })
          .then(res0 => {
            console.log(
              `deleted all shards in ${indexname}: ${JSON.stringify(
                res0
              )}`
            )
          })
          .catch(err => {
            const errmessage = `error deleting shards in index ${indexname}: ${err}`
            console.log(errmessage)
          })
          .then(() => {
            writeclient.indices
              .delete({
                index: indexname
              })
              .then(res1 => {
                console.log(
                  `deleted index ${indexname}: ${JSON.stringify(res1)}`
                )
              })
              .catch(err => {
                const errmessage = `error deleting index ${indexname}: ${err}`
                console.log(errmessage)
              })
              .then(() => {
                return writeclient.indices
                  .exists({
                    index: indexname
                  })
                  .then(res2 => {
                    if (res2) {
                      console.log(`index ${indexname} exists still`)
                    } else {
                      return writeclient.indices
                        .create({
                          index: indexname,
                          body: {
                            settings: indexsettings
                          }
                        })
                        .then(res3 => {
                          console.log(`added index ${indexname}: ${res3}`)
                          return writeclient.indices
                            .getMapping()
                            .then(res4 => {
                              if (
                                Object.keys(res4[indexname].mappings)
                                  .length === 0
                              ) {
                                console.log(
                                  `${indexname}: no mappings :)`
                                )
                                return writeclient.indices
                                  .putMapping({
                                    index: indexname,
                                    type: doctype,
                                    body: mappings,
                                    include_type_name: true
                                  })
                                  .then(res5 => {
                                    console.log(
                                      `initialized ${indexname}: ${JSON.stringify(
                                        res5
                                      )}`
                                    )
                                    resolve(`finished initializing elasticsearch`)
                                  })
                                  .catch(err => {
                                    const errmessage = `could not create mapping for ${indexname}: ${err}`
                                    console.log(errmessage)
                                    reject(errmessage)
                                  })
                              } else {
                                const errmessage = `${indexname} already has mappings :(`
                                console.log(errmessage)
                                reject(errmessage)
                              }
                            })
                            .catch(err => {
                              const errmessage = `could not get mappings for ${indexname}: ${err}`
                              console.log(errmessage)
                              reject(errmessage)
                            })
                        })

                        .catch(err => {
                          const errmessage = `error adding index ${indexname}: ${err}`
                          console.log(errmessage)
                          reject(errmessage)
                        })
                    }
                  })
                  .catch(err => {
                    const errmessage = `error checking if index ${indexname} exists: ${err}`
                    console.log(errmessage)
                    reject(errmessage)
                  })
              })
          })
      })
      .catch(err => {
        const errmessage = `unable to ping writeclient`
        console.log(errmessage)
        reject(err)
      })
  })
}
