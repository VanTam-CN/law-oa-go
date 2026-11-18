import request from '../utils/request';

export async function getUserList(params) {
  return request({
    url: '/api/v1/admin/users',
    method: 'get',
    params
  });
}

export async function getUserDetail(userId) {
  return request({
    url: `/api/v1/admin/users/${userId}`,
    method: 'get'
  });
}

export async function createUser(data) {
  return request({
    url: '/api/v1/admin/users',
    method: 'post',
    data
  });
}

export async function updateUser(userId, data) {
  return request({
    url: `/api/v1/admin/users/${userId}`,
    method: 'put',
    data
  });
}

export async function deleteUser(userId) {
  return request({
    url: `/api/v1/admin/users/${userId}`,
    method: 'delete'
  });
}

// 基础用户管理功能
export async function changeUserStatus(userId, status) {
  return request({
    url: `/api/v1/admin/users/${userId}/status`,
    method: 'put',
    data: { status }
  });
}

// 用户注册使用 auth 模块的 register 函数

// 用户配置文件管理已在 auth 模块中实现

// 以下函数可能需要根据后端API调整或暂时注释掉
/*
export async function getUserRoleIds(userId) {
  return request({
    url: `/api/v1/admin/users/${userId}/roles`,
    method: 'get'
  });
}

export async function updateUserRole(userId, roleIds) {
  return request({
    url: `/api/v1/admin/users/${userId}/roles`,
    method: 'put',
    data: { roleIds }
  });
}

export async function getUserOptions() {
  return request({
    url: '/api/v1/users/options',
    method: 'get'
  });
}

export async function exportUsers(params) {
  return request({
    url: '/api/v1/admin/users/export',
    method: 'post',
    params,
    responseType: 'blob'
  });
}

export async function importUsers(data, updateSupport) {
  return request({
    url: `/api/v1/admin/users/import?updateSupport=${updateSupport}`,
    method: 'post',
    data,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  });
}

export async function downloadUserTemplate() {
  return request({
    url: '/api/v1/admin/users/template',
    method: 'get',
    responseType: 'blob'
  });
}
*/