import { useState } from 'react'
import { Search, Plus, Filter, CheckCircle, Circle, AlertTriangle, Calendar, User, X } from 'lucide-react'
import { useTODOs, useTODOMutations } from '../../hooks/useAPI'

interface TODO {
  id: string
  title: string
  description: string
  priority: number
  status: 'pending' | 'in-progress' | 'completed'
  assignee?: string
  dueDate?: string
  createdAt: string
}

interface CreateTODOModalProps {
  isOpen: boolean
  onClose: () => void
  onSubmit: (data: CreateTODORequest) => Promise<void>
}

interface CreateTODORequest {
  title: string
  description: string
  priority: number
}

function CreateTODOModal({ isOpen, onClose, onSubmit }: CreateTODOModalProps) {
  const [formData, setFormData] = useState<CreateTODORequest>({
    title: '',
    description: '',
    priority: 5,
  })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!formData.title.trim()) {
      setError('Title is required')
      return
    }
    
    setIsSubmitting(true)
    setError(null)
    
    try {
      await onSubmit(formData)
      setFormData({ title: '', description: '', priority: 5 })
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create TODO')
    } finally {
      setIsSubmitting(false)
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm animate-fade-in">
      <div className="bg-surface border border-border rounded-lg shadow-xl w-full max-w-md mx-4">
        <div className="flex items-center justify-between p-4 border-b border-border">
          <h2 className="text-xl font-semibold text-text-primary">Create New TODO</h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-surface-hover rounded-lg text-text-muted hover:text-text-primary transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>
        
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {error && (
            <div className="p-3 bg-error/10 border border-error/20 rounded-lg text-error text-sm">
              {error}
            </div>
          )}
          
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              Title
            </label>
            <input
              type="text"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              placeholder="e.g., Update API documentation"
              className="input w-full"
              autoFocus
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              Description
            </label>
            <textarea
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              placeholder="Add details about this task..."
              className="input w-full min-h-[100px] resize-y"
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              Priority (1-10): {formData.priority}
            </label>
            <input
              type="range"
              min="1"
              max="10"
              value={formData.priority}
              onChange={(e) => setFormData({ ...formData, priority: parseInt(e.target.value) })}
              className="w-full h-2 bg-surface-hover rounded-lg appearance-none cursor-pointer"
            />
            <div className="flex justify-between text-xs text-text-muted mt-1">
              <span>Low</span>
              <span>Medium</span>
              <span>High</span>
            </div>
          </div>
          
          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="btn-secondary flex-1"
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn-primary flex-1"
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Creating...' : 'Create TODO'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function TODOList() {
  const [search, setSearch] = useState('')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [filterPriority, setFilterPriority] = useState<string>('all')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [toast, setToast] = useState<{ type: 'success' | 'error'; message: string } | null>(null)

  const { data: todos, isLoading, error } = useTODOs()
  const mutations = useTODOMutations()

  const showToast = (type: 'success' | 'error', message: string) => {
    setToast({ type, message })
    setTimeout(() => setToast(null), 3000)
  }

  const handleCreateTODO = async (data: CreateTODORequest) => {
    await mutations.create.mutateAsync(data)
    showToast('success', `TODO "${data.title}" created successfully!`)
  }

  const handleCompleteTODO = async (id: string, title: string) => {
    try {
      await mutations.complete.mutateAsync(id)
      showToast('success', `TODO "${title}" completed!`)
    } catch (err) {
      showToast('error', 'Failed to complete TODO')
    }
  }

  const handleDeleteTODO = async (id: string, title: string) => {
    if (confirm(`Are you sure you want to delete "${title}"?`)) {
      try {
        await mutations.delete.mutateAsync(id)
        showToast('success', `TODO "${title}" deleted successfully!`)
      } catch (err) {
        showToast('error', 'Failed to delete TODO')
      }
    }
  }

  const getPriorityLabel = (priority: number): string => {
    if (priority <= 3) return 'low'
    if (priority <= 7) return 'medium'
    return 'high'
  }

  const filteredTODOs = (todos || []).filter((todo: TODO) => {
    const matchesSearch = todo.title.toLowerCase().includes(search.toLowerCase())
    const matchesStatus = filterStatus === 'all' || todo.status === filterStatus
    const todoPriorityLabel = getPriorityLabel(todo.priority)
    const matchesPriority = filterPriority === 'all' || todoPriorityLabel === filterPriority
    return matchesSearch && matchesStatus && matchesPriority
  })

  if (error) {
    return (
      <div className="card text-center py-12">
        <p className="text-error">Failed to load TODOs. Please check if the backend is running.</p>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Toast Notification */}
      {toast && (
        <div className={`fixed top-4 right-4 z-50 p-4 rounded-lg shadow-lg animate-fade-in ${
          toast.type === 'success' ? 'bg-success/20 border border-success/30 text-success' : 'bg-error/20 border border-error/30 text-error'
        }`}>
          {toast.message}
        </div>
      )}

      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">TODOs</h1>
          <p className="text-text-secondary mt-1">Track and manage your tasks</p>
        </div>
        <button 
          onClick={() => setShowCreateModal(true)}
          className="btn-primary"
        >
          <Plus className="w-4 h-4 mr-2" />
          New TODO
        </button>
      </div>

      {/* Search and Filters */}
      <div className="flex flex-wrap gap-4">
        <div className="relative flex-1 min-w-64">
          <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" />
          <input
            type="text"
            placeholder="Search TODOs..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="input pl-10"
          />
        </div>
        <div className="flex items-center gap-2">
          <Filter className="w-4 h-4 text-text-muted" />
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="input w-auto"
          >
            <option value="all">All Status</option>
            <option value="pending">Pending</option>
            <option value="in-progress">In Progress</option>
            <option value="completed">Completed</option>
          </select>
          <select
            value={filterPriority}
            onChange={(e) => setFilterPriority(e.target.value)}
            className="input w-auto"
          >
            <option value="all">All Priority</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>
        </div>
      </div>

      {/* Loading State */}
      {isLoading && (
        <div className="card text-center py-12">
          <p className="text-text-muted">Loading TODOs...</p>
        </div>
      )}

      {/* TODO List */}
      {!isLoading && (
        <div className="space-y-3">
          {filteredTODOs.map((todo: TODO) => (
            <TODOCard 
              key={todo.id} 
              todo={todo}
              onComplete={() => handleCompleteTODO(todo.id, todo.title)}
              onDelete={() => handleDeleteTODO(todo.id, todo.title)}
            />
          ))}
        </div>
      )}

      {!isLoading && filteredTODOs.length === 0 && (
        <div className="card text-center py-12">
          <p className="text-text-muted">No TODOs found</p>
          <button 
            onClick={() => setShowCreateModal(true)}
            className="btn-primary mt-4"
          >
            Create your first TODO
          </button>
        </div>
      )}

      {/* Create TODO Modal */}
      <CreateTODOModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSubmit={handleCreateTODO}
      />
    </div>
  )
}

function TODOCard({ todo, onComplete, onDelete }: { todo: TODO; onComplete: () => void; onDelete: () => void }) {
  const priorityColors: Record<string, string> = {
    high: 'bg-error/20 text-error',
    medium: 'bg-warning/20 text-warning',
    low: 'bg-accent/20 text-accent',
  }

  const statusIcons: Record<string, typeof Circle> = {
    pending: Circle,
    'in-progress': AlertTriangle,
    completed: CheckCircle,
  }

  const StatusIcon = statusIcons[todo.status]
  const priorityLabel = todo.priority <= 3 ? 'low' : todo.priority <= 7 ? 'medium' : 'high'

  return (
    <div className={`card group hover:border-border-light transition-colors ${todo.status === 'completed' ? 'opacity-60' : ''}`}>
      <div className="flex items-start gap-4">
        {/* Status checkbox */}
        <button 
          onClick={onComplete}
          className={`mt-1 ${todo.status === 'completed' ? 'text-success' : 'text-text-muted hover:text-accent'}`}
        >
          <StatusIcon className="w-5 h-5" />
        </button>

        {/* Content */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h3 className={`font-medium ${todo.status === 'completed' ? 'line-through text-text-muted' : 'text-text-primary'}`}>
                {todo.title}
              </h3>
              <p className="text-sm text-text-muted mt-1">{todo.description}</p>
            </div>
            <div className="flex items-center gap-2">
              <span className={`badge ${priorityColors[priorityLabel]}`}>
                {priorityLabel} ({todo.priority})
              </span>
            </div>
          </div>

          {/* Meta info */}
          <div className="flex flex-wrap items-center gap-4 mt-3 text-sm text-text-muted">
            {todo.assignee && (
              <div className="flex items-center gap-1">
                <User className="w-3.5 h-3.5" />
                <span>{todo.assignee}</span>
              </div>
            )}
            {todo.dueDate && (
              <div className="flex items-center gap-1">
                <Calendar className="w-3.5 h-3.5" />
                <span>{todo.dueDate}</span>
              </div>
            )}
            <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
              <button className="text-text-muted hover:text-accent">Edit</button>
              <span>•</span>
              <button 
                onClick={onDelete}
                className="text-text-muted hover:text-error"
              >
                Delete
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
