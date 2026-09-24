import axios from 'axios'

export const API_BASE = import.meta.env.VITE_API_BASE || '/api'
const api = axios.create({ baseURL: API_BASE, withCredentials: true, timeout: 12000 })
api.interceptors.response.use(response => response, error => {
  if (error.response?.status === 401 && !error.config?.url?.includes('verify-otp')) {
    window.dispatchEvent(new Event('session-expired'))
  }
  return Promise.reject(error)
})
export function errorMessage(error: unknown): string {
  return axios.isAxiosError(error)
    ? error.response?.data?.error || 'Could not reach UnAlone. Please try again.'
    : 'Something went wrong. Please try again.'
}
export default api
