import * as rpc from "vlens/rpc"

export interface AddUserRequest {
    Username: string
    Email: string
    Password: string
}

export interface UserListResponse {
    Users: User[]
}

export interface Empty {
}

export interface User {
    Id: number
    Username: string
    Email: string
    IsAdmin: boolean
}

export async function AddUser(data: AddUserRequest): Promise<rpc.Response<UserListResponse>> {
    return await rpc.call<UserListResponse>('AddUser', JSON.stringify(data));
}

export async function ListUsers(data: Empty): Promise<rpc.Response<UserListResponse>> {
    return await rpc.call<UserListResponse>('ListUsers', JSON.stringify(data));
}

