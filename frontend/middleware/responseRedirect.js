const responseRedirect = ({ redirect, params }) => {
  redirect(`/project/${params.projectId}/form/${params.formId}/view`)
}

export default responseRedirect
