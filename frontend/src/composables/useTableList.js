import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'

export function useTableList(fetchApi, deleteApi) {
  const loading = ref(false)
  const tableData = ref([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(10)
  const keyword = ref('')

  const fetchData = async () => {
    loading.value = true
    try {
      const res = await fetchApi({
        page: page.value,
        pageSize: pageSize.value,
        keyword: keyword.value
      })
      tableData.value = res.data.list || []
      total.value = res.data.total || 0
    } catch (error) {
      ElMessage.error('获取数据失败')
    } finally {
      loading.value = false
    }
  }

  const handlePageChange = (val) => {
    page.value = val
    fetchData()
  }

  const handleSearch = () => {
    page.value = 1
    fetchData()
  }

  const handleDelete = async (id, name = '此项') => {
    try {
      await ElMessageBox.confirm(`确定删除${name}吗？`, '提示', {
        type: 'warning'
      })
      await deleteApi(id)
      ElMessage.success('删除成功')
      fetchData()
    } catch (error) {
      if (error !== 'cancel') {
        ElMessage.error('删除失败')
      }
    }
  }

  return {
    loading,
    tableData,
    total,
    page,
    pageSize,
    keyword,
    fetchData,
    handlePageChange,
    handleSearch,
    handleDelete
  }
}
