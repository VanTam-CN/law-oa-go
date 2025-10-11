import { get, post, put, del } from '../services/http';
import { userService } from '../services/user';

export { userService };

// 导出具体函数以保持向后兼容
export const getUsers = userService.getUserList;
export const getUser = userService.getCurrentUser;
export const createUser = (data: any) => post('/users', data);
export const updateUser = (id: number, data: any) => put(`/users/${id}`, data);
export const deleteUser = (id: number) => del(`/users/${id}`);