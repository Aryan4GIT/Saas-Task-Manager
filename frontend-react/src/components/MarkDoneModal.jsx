import { useState } from 'react';
import { taskService } from '../services/api.service';
import toast from 'react-hot-toast';
import { X, Upload, FileText } from 'lucide-react';

export default function MarkDoneModal({ task, onClose }) {
  const [file, setFile] = useState(null);
  const [uploading, setUploading] = useState(false);

  const handleFileChange = (e) => {
    const selectedFile = e.target.files[0];
    if (selectedFile) {
      // Check file size (limit to 10MB)
      if (selectedFile.size > 10 * 1024 * 1024) {
        toast.error('File size must be less than 10MB');
        return;
      }
      setFile(selectedFile);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!file) {
      toast.error('Please select a document to upload');
      return;
    }

    setUploading(true);

    try {
      const formData = new FormData();
      formData.append('document', file);

      await taskService.markDone(task.id, formData);
      toast.success('Task marked as done with document uploaded!');
      onClose(true);
    } catch (error) {
      const message = error.response?.data?.message || 'Failed to mark task as done';
      toast.error(message);
    } finally {
      setUploading(false);
    }
  };

  const handleSkip = async () => {
    if (!confirm('Mark as done without uploading a document?')) return;
    
    setUploading(true);
    try {
      await taskService.markDone(task.id);
      toast.success('Task marked as done!');
      onClose(true);
    } catch (error) {
      toast.error('Failed to mark task as done');
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4 z-50">
      <div className="rounded-2xl w-full max-w-md bg-slate-900 border border-slate-800 shadow-2xl">
        <div className="px-6 py-4 flex justify-between items-center bg-slate-800 border-b border-slate-700">
          <h2 className="text-2xl font-bold text-white">Mark Task as Done</h2>
          <button
            onClick={() => onClose(false)}
            className="p-2 hover:bg-slate-700 rounded-lg transition-colors"
          >
            <X className="w-5 h-5 text-gray-400" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <p className="text-sm text-blue-900">
              <strong>📄 Upload your completed work document</strong><br />
              This helps managers and admins review your work faster with AI-generated summaries.
            </p>
          </div>

          <div className="space-y-2">
            <label className="block text-sm font-medium text-gray-700">
              Document <span className="text-gray-500">(PDF, DOC, TXT, etc.)</span>
            </label>
            
            <div className="relative">
              <input
                type="file"
                id="document"
                onChange={handleFileChange}
                className="hidden"
                accept=".pdf,.doc,.docx,.txt,.md,.log,.csv,.xlsx,.xls"
              />
              <label
                htmlFor="document"
                className="flex items-center justify-center gap-3 px-4 py-8 border-2 border-dashed border-gray-300 rounded-lg cursor-pointer hover:border-primary-500 hover:bg-gray-50 transition-colors"
              >
                <Upload className="w-8 h-8 text-gray-400" />
                <div className="text-center">
                  <p className="text-sm font-medium text-gray-700">
                    {file ? file.name : 'Click to upload document'}
                  </p>
                  <p className="text-xs text-gray-500 mt-1">
                    {file ? `${(file.size / 1024).toFixed(2)} KB` : 'Max 10MB'}
                  </p>
                </div>
              </label>
            </div>

            {file && (
              <div className="flex items-center gap-2 p-3 bg-green-50 border border-green-200 rounded-lg">
                <FileText className="w-5 h-5 text-green-600" />
                <span className="text-sm text-green-900 flex-1">{file.name}</span>
                <button
                  type="button"
                  onClick={() => setFile(null)}
                  className="text-red-600 hover:text-red-700"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            )}
          </div>

          <div className="flex gap-3 pt-4">
            <button
              type="submit"
              disabled={uploading || !file}
              className="flex-1 btn btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {uploading ? 'Uploading...' : 'Upload & Mark Done'}
            </button>
            <button
              type="button"
              onClick={handleSkip}
              disabled={uploading}
              className="flex-1 btn btn-secondary"
            >
              Skip & Mark Done
            </button>
          </div>

          <p className="text-xs text-gray-500 text-center">
            AI will generate a summary of your document to help managers review faster
          </p>
        </form>
      </div>
    </div>
  );
}
