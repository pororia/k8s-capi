export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: {
    code: string
    message: string
  }
}

export interface PaginatedResponse<T> {
  success: boolean
  data: T[]
  pagination: {
    page: number
    page_size: number
    total: number
    total_pages: number
  }
}
