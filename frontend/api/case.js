import request from '../utils/request';

export async function getCaseList(params) {
  return request({
    url: '/api/v1/cases',
    method: 'get',
    params
  });
}

export async function getCaseDetail(caseId) {
  return request({
    url: `/api/v1/cases/${caseId}`,
    method: 'get'
  });
}

export async function createCase(data) {
  return request({
    url: '/api/v1/cases',
    method: 'post',
    data
  });
}

export async function updateCase(caseId, data) {
  return request({
    url: `/api/v1/cases/${caseId}`,
    method: 'put',
    data
  });
}

export async function deleteCase(caseId) {
  return request({
    url: `/api/v1/cases/${caseId}`,
    method: 'delete'
  });
}

export async function getCaseStats() {
  return request({
    url: '/api/v1/cases/stats',
    method: 'get'
  });
}

export async function assignLawyer(caseId, lawyerId) {
  return request({
    url: `/api/v1/cases/${caseId}/assign`,
    method: 'post',
    data: { lawyer_id: lawyerId }
  });
}

export async function updateCaseStatus(caseId, status) {
  return request({
    url: `/api/v1/cases/${caseId}/status`,
    method: 'post',
    data: { status }
  });
}

export async function searchCases(params) {
  return request({
    url: '/api/v1/cases',
    method: 'get',
    params: {
      ...params,
      search: params.keyword || params.search
    }
  });
}

export async function getCasesByType(caseType) {
  return request({
    url: '/api/v1/cases',
    method: 'get',
    params: { case_type: caseType }
  });
}

export async function getCasesByStatus(status) {
  return request({
    url: '/api/v1/cases',
    method: 'get',
    params: { status }
  });
}

export async function getCasesByPriority(priority) {
  return request({
    url: '/api/v1/cases',
    method: 'get',
    params: { priority }
  });
}

export async function getCasesByClient(clientId) {
  return request({
    url: '/api/v1/cases',
    method: 'get',
    params: { client_id: clientId }
  });
}

export async function getCasesByLawyer(lawyerId) {
  return request({
    url: '/api/v1/cases',
    method: 'get',
    params: { lawyer_id: lawyerId }
  });
}

export async function bulkUpdateCases(data) {
  return request({
    url: '/api/v1/cases/bulk',
    method: 'put',
    data
  });
}

export async function exportCases(params) {
  return request({
    url: '/api/v1/cases/export',
    method: 'post',
    params,
    responseType: 'blob'
  });
}