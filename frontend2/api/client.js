import request from '../utils/request';

export async function getClientList(params) {
  return request({
    url: '/api/v1/clients',
    method: 'get',
    params
  });
}

export async function getClientDetail(clientId) {
  return request({
    url: `/api/v1/clients/${clientId}`,
    method: 'get'
  });
}

export async function createClient(data) {
  return request({
    url: '/api/v1/clients',
    method: 'post',
    data
  });
}

export async function updateClient(clientId, data) {
  return request({
    url: `/api/v1/clients/${clientId}`,
    method: 'put',
    data
  });
}

export async function deleteClient(clientId) {
  return request({
    url: `/api/v1/clients/${clientId}`,
    method: 'delete'
  });
}

export async function getClientStats() {
  return request({
    url: '/api/v1/clients/stats',
    method: 'get'
  });
}

export async function searchClients(params) {
  return request({
    url: '/api/v1/clients',
    method: 'get',
    params: {
      ...params,
      search: params.keyword || params.search
    }
  });
}

export async function getClientsByStatus(status) {
  return request({
    url: '/api/v1/clients',
    method: 'get',
    params: { status }
  });
}

export async function bulkUpdateClients(data) {
  return request({
    url: '/api/v1/clients/bulk',
    method: 'put',
    data
  });
}

export async function exportClients(params) {
  return request({
    url: '/api/v1/clients/export',
    method: 'post',
    params,
    responseType: 'blob'
  });
}