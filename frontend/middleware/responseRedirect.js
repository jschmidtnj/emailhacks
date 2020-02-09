const responseRedirect = ({ redirect, params }) => {
  redirect(`/project/${params.project}/form/${params.formId}/view`)
}

export default responseRedirect
