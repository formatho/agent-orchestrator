import { useState } from 'react'
import { Search, Plus, Play, Pause, Clock, Trash2, Edit, X } from 'lucide-react'
import { useCronJobs, useCronMutations, useAgents } from '../../hooks/useAPI'

interface CronJob {
  id: string
  name: string
  schedule: string
  scheduleHuman?: string
  command?: string
  agent_id?: string
  status: 'active' | 'paused' | 'error'
  lastRun?: string
  nextRun: string
  successCount: number
  failCount: number
}

interface CreateCronModalProps {
  isOpen: boolean
  onClose: () => void
  onSubmit: (data: CreateCronRequest) => Promise<void>
  agents: Array<{ id: string; name: string }>
}

interface CreateCronRequest {
  name: string
  schedule: string
  agent_id: string
}

function CreateCronModal({ isOpen, onClose, onSubmit, agents }: CreateCronModalProps) {
  const [formData, setFormData] = useState<CreateCronRequest>({
    name: '',
    schedule: '0 9 * * *',
    agent_id: '',
  })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const schedulePresets = [
    { label: 'Every hour', value: '0 * * * *' },
    { label: 'Every day at 9 AM', value: '0 9 * * *' },
    { label: 'Every day at midnight', value: '0 0 * * *' },
    { label: 'Every Sunday at midnight', value: '0 0 * * 0' },
    { label: 'Every 5 minutes', value: '*/5 * * * *' },
    { label: 'Every 30 minutes', value: '*/30 * * * *' },
  ]

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!formData.name.trim()) {
      setError('Job name is required')
      return
    }
    if (!formData.schedule.trim()) {
      setError('Schedule is required')
      return
    }
    
    setIsSubmitting(true)
    setError(null)
    
    try {
      await onSubmit(formData)
      setFormData({ name: '', schedule: '0 9 * * *', agent_id: '' })
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create cron job')
    } finally {
      setIsSubmitting(false)
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm animate-fade-in">
      <div className="bg-surface border border-border rounded-lg shadow-xl w-full max-w-md mx-4">
        <div className="flex items-center justify-between p-4 border-b border-border">
          <h2 className="text-xl font-semibold text-text-primary">Create New Cron Job</h2>
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
              Job Name
            </label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="e.g., daily-report"
              className="input w-full"
              autoFocus
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              Schedule (Cron Syntax)
            </label>
            <input
              type="text"
              value={formData.schedule}
              onChange={(e) => setFormData({ ...formData, schedule: e.target.value })}
              placeholder="* * * * *"
              className="input w-full font-mono"
            />
            <div className="mt-2 flex flex-wrap gap-2">
              {schedulePresets.map((preset) => (
                <button
                  key={preset.value}
                  type="button"
                  onClick={() => setFormData({ ...formData, schedule: preset.value })}
                  className="text-xs px-2 py-1 bg-surface-hover hover:bg-border rounded text-text-muted hover:text-text-primary transition-colors"
                >
                  {preset.label}
                </button>
              ))}
            </div>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              Assign to Agent
            </label>
            <select
              value={formData.agent_id}
              onChange={(e) => setFormData({ ...formData, agent_id: e.target.value })}
              className="input w-full"
            >
              <option value="">No agent assigned</option>
              {agents.map((agent) => (
                <option key={agent.id} value={agent.id}>
                  {agent.name}
                </option>
              ))}
            </select>
            <p className="text-xs text-text-muted mt-1">Optional: Assign this cron job to an agent</p>
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
              {isSubmitting ? 'Creating...' : 'Create Cron Job'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function CronList() {
  const [search, setSearch] = useState('')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [toast, setToast] = useState<{ type: 'success' | 'error'; message: string } | null>(null)

  const { data: cronJobs, isLoading: jobsLoading, error: jobsError } = useCronJobs()
  const { data: agents } = useAgents()
  const mutations = useCronMutations()

  const showToast = (type: 'success' | 'error', message: string) => {
    setToast({ type, message })
    setTimeout(() => setToast(null), 3000)
  }

  const handleCreateCron = async (data: CreateCronRequest) => {
    await mutations.create.mutateAsync(data)
    showToast('success', `Cron job "${data.name}" created successfully!`)
  }

  const handleToggleJob = async (job: CronJob) => {
    try {
      if (job.status === 'active') {
        await mutations.pause.mutateAsync(job.id)
        showToast('success', `Cron job "${job.name}" paused`)
      } else {
        await mutations.resume.mutateAsync(job.id)
        showToast('success', `Cron job "${job.name}" resumed`)
      }
    } catch (err) {
      showToast('error', 'Failed to toggle cron job')
    }
  }

  const handleDeleteJob = async (id: string, name: string) => {
    if (confirm(`Are you sure you want to delete cron job "${name}"?`)) {
      try {
        await mutations.delete.mutateAsync(id)
        showToast('success', `Cron job "${name}" deleted successfully!`)
      } catch (err) {
        showToast('error', 'Failed to delete cron job')
      }
    }
  }

  const filteredJobs = (cronJobs || []).filter((job: CronJob) =>
    job.name.toLowerCase().includes(search.toLowerCase())
  )

  if (jobsError) {
    return (
      <div className="card text-center py-12">
        <p className="text-error">Failed to load cron jobs. Please check if the backend is running.</p>
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
          <h1 className="text-2xl font-bold text-text-primary">Cron Jobs</h1>
          <p className="text-text-secondary mt-1">Schedule and automate recurring tasks</p>
        </div>
        <button 
          onClick={() => setShowCreateModal(true)}
          className="btn-primary"
        >
          <Plus className="w-4 h-4 mr-2" />
          New Cron Job
        </button>
      </div>

      {/* Search */}
      <div className="relative max-w-md">
        <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" />
        <input
          type="text"
          placeholder="Search cron jobs..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="input pl-10"
        />
      </div>

      {/* Loading State */}
      {jobsLoading && (
        <div className="card text-center py-12">
          <p className="text-text-muted">Loading cron jobs...</p>
        </div>
      )}

      {/* Cron Jobs Table */}
      {!jobsLoading && (
        <div className="card overflow-hidden p-0">
          <table className="w-full">
            <thead>
              <tr className="border-b border-border">
                <th className="text-left p-4 text-sm font-medium text-text-secondary">Name</th>
                <th className="text-left p-4 text-sm font-medium text-text-secondary hidden md:table-cell">Schedule</th>
                <th className="text-left p-4 text-sm font-medium text-text-secondary hidden lg:table-cell">Agent</th>
                <th className="text-left p-4 text-sm font-medium text-text-secondary">Status</th>
                <th className="text-left p-4 text-sm font-medium text-text-secondary hidden sm:table-cell">Next Run</th>
                <th className="text-right p-4 text-sm font-medium text-text-secondary">Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredJobs.map((job: CronJob) => (
                <tr key={job.id} className="border-b border-border last:border-0 hover:bg-surface-hover transition-colors">
                  <td className="p-4">
                    <div>
                      <p className="font-medium text-text-primary">{job.name}</p>
                      <p className="text-xs text-text-muted md:hidden">{job.schedule}</p>
                    </div>
                  </td>
                  <td className="p-4 hidden md:table-cell">
                    <div className="flex items-center gap-2">
                      <Clock className="w-4 h-4 text-text-muted" />
                      <div>
                        <code className="text-sm text-accent">{job.schedule}</code>
                        {job.scheduleHuman && (
                          <p className="text-xs text-text-muted">{job.scheduleHuman}</p>
                        )}
                      </div>
                    </div>
                  </td>
                  <td className="p-4 hidden lg:table-cell">
                    {job.agent_id ? (
                      <span className="text-sm text-text-secondary">
                        {agents?.find((a: any) => a.id === job.agent_id)?.name || job.agent_id}
                      </span>
                    ) : (
                      <span className="text-sm text-text-muted">-</span>
                    )}
                  </td>
                  <td className="p-4">
                    <div className="flex items-center gap-2">
                      <span className={`status-dot ${
                        job.status === 'active' ? 'online' :
                        job.status === 'error' ? 'error' :
                        'offline'
                      }`} />
                      <span className="text-sm capitalize">{job.status}</span>
                    </div>
                  </td>
                  <td className="p-4 hidden sm:table-cell">
                    <span className="text-sm text-text-secondary">{job.nextRun}</span>
                  </td>
                  <td className="p-4">
                    <div className="flex items-center justify-end gap-1">
                      <button 
                        onClick={() => handleToggleJob(job)}
                        className="p-2 hover:bg-surface rounded-lg text-text-muted hover:text-text-primary" 
                        title={job.status === 'active' ? 'Pause' : 'Resume'}
                      >
                        {job.status === 'active' ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
                      </button>
                      <button className="p-2 hover:bg-surface rounded-lg text-text-muted hover:text-text-primary" title="Edit">
                        <Edit className="w-4 h-4" />
                      </button>
                      <button 
                        onClick={() => handleDeleteJob(job.id, job.name)}
                        className="p-2 hover:bg-surface rounded-lg text-text-muted hover:text-error" 
                        title="Delete"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>

          {filteredJobs.length === 0 && (
            <div className="text-center py-12">
              <p className="text-text-muted">No cron jobs found</p>
              <button 
                onClick={() => setShowCreateModal(true)}
                className="btn-primary mt-4"
              >
                Create your first cron job
              </button>
            </div>
          )}
        </div>
      )}

      {/* Stats Summary */}
      {!jobsLoading && cronJobs && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="card">
            <p className="text-sm text-text-secondary">Total Jobs</p>
            <p className="text-2xl font-bold mt-1">{cronJobs.length}</p>
          </div>
          <div className="card">
            <p className="text-sm text-text-secondary">Active Jobs</p>
            <p className="text-2xl font-bold mt-1 text-success">{cronJobs.filter((j: CronJob) => j.status === 'active').length}</p>
          </div>
          <div className="card">
            <p className="text-sm text-text-secondary">Total Executions</p>
            <p className="text-2xl font-bold mt-1">
              {cronJobs.reduce((acc: number, j: CronJob) => acc + j.successCount + j.failCount, 0).toLocaleString()}
            </p>
          </div>
        </div>
      )}

      {/* Create Cron Job Modal */}
      <CreateCronModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSubmit={handleCreateCron}
        agents={agents || []}
      />
    </div>
  )
}
