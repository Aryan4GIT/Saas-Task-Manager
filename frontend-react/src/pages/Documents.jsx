import { useEffect, useMemo, useState } from 'react';
import toast from 'react-hot-toast';
import { FileText, UploadCloud, Search, CheckCircle, XCircle, Clock, Sparkles } from 'lucide-react';
import { documentService } from '../services/api.service';
import { useAuth } from '../context/AuthContext';

export default function Documents() {
  const { user } = useAuth();
  const isAdminOrManager = user?.role === 'admin' || user?.role === 'manager';

  const [docs, setDocs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [viewMode, setViewMode] = useState('all'); // 'all' | 'pending'

  const [file, setFile] = useState(null);
  const [title, setTitle] = useState('');
  const [uploading, setUploading] = useState(false);

  const [selectedId, setSelectedId] = useState('');
  const selectedDoc = useMemo(() => docs.find((d) => d.id === selectedId) || null, [docs, selectedId]);

  const [question, setQuestion] = useState('');
  const [verifying, setVerifying] = useState(false);
  const [result, setResult] = useState(null);

  // AI Summary
  const [summary, setSummary] = useState(null);
  const [summarizing, setSummarizing] = useState(false);

  // Status update
  const [statusNotes, setStatusNotes] = useState('');
  const [updating, setUpdating] = useState(false);

  useEffect(() => {
    loadDocs();
  }, [viewMode]);

  const loadDocs = async () => {
    try {
      setLoading(true);
      let res;
      if (viewMode === 'pending' && isAdminOrManager) {
        res = await documentService.listPendingDocuments();
      } else {
        res = await documentService.listDocuments();
      }
      const arr = Array.isArray(res.data) ? res.data : (res.data?.data || []);
      setDocs(arr);
      if (!selectedId && arr.length > 0) {
        setSelectedId(arr[0].id);
      }
    } catch (e) {
      console.error('[Documents] Failed to load:', e);
      toast.error('Failed to load documents');
    } finally {
      setLoading(false);
    }
  };

  const onUpload = async (e) => {
    e.preventDefault();
    if (!file) {
      toast.error('Choose a file to upload');
      return;
    }

    try {
      setUploading(true);
      const res = await documentService.uploadDocument(file, title.trim() || undefined);
      toast.success('Uploaded successfully');
      const created = res.data;
      await loadDocs();
      if (created?.id) {
        setSelectedId(created.id);
      }
      setFile(null);
      setTitle('');
      setResult(null);
      setSummary(null);
    } catch (e2) {
      const msg = e2?.response?.data?.message || e2?.response?.data?.error || 'Upload failed';
      toast.error(msg);
    } finally {
      setUploading(false);
    }
  };

  const onVerify = async (e) => {
    e.preventDefault();
    if (!selectedDoc) {
      toast.error('Select a document first');
      return;
    }
    if (!question.trim()) {
      toast.error('Enter a verification question');
      return;
    }

    try {
      setVerifying(true);
      setResult(null);
      const res = await documentService.verifyDocument(selectedDoc.id, question.trim());
      setResult(res.data);
    } catch (e2) {
      const msg = e2?.response?.data?.message || e2?.response?.data?.error || 'Verification failed';
      toast.error(msg);
    } finally {
      setVerifying(false);
    }
  };

  const onGenerateSummary = async () => {
    if (!selectedDoc) return;

    try {
      setSummarizing(true);
      setSummary(null);
      const res = await documentService.generateSummary(selectedDoc.id);
      setSummary(res.data);
    } catch (e2) {
      const msg = e2?.response?.data?.message || e2?.response?.data?.error || 'Summary generation failed';
      toast.error(msg);
    } finally {
      setSummarizing(false);
    }
  };

  const onUpdateStatus = async (status) => {
    if (!selectedDoc) return;

    try {
      setUpdating(true);
      await documentService.updateDocumentStatus(selectedDoc.id, status, statusNotes);
      toast.success(`Document ${status}`);
      setStatusNotes('');
      await loadDocs();
    } catch (e2) {
      const msg = e2?.response?.data?.message || e2?.response?.data?.error || 'Status update failed';
      toast.error(msg);
    } finally {
      setUpdating(false);
    }
  };

  const verdictBadge = (verdict) => {
    const v = String(verdict || '').toLowerCase();
    if (v === 'verified') return 'bg-emerald-100 text-emerald-800';
    if (v === 'unverified') return 'bg-red-100 text-red-800';
    return 'bg-yellow-100 text-yellow-800';
  };

  const statusBadge = (status) => {
    const s = String(status || 'submitted').toLowerCase();
    if (s === 'verified') return { bg: 'bg-emerald-100 text-emerald-800', icon: CheckCircle };
    if (s === 'rejected') return { bg: 'bg-red-100 text-red-800', icon: XCircle };
    return { bg: 'bg-yellow-100 text-yellow-800', icon: Clock };
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h2 className="text-3xl font-bold text-white">Documents</h2>
          <p className="text-white/80 text-sm mt-1">Upload work documents for verification and AI analysis</p>
        </div>
        {isAdminOrManager && (
          <div className="flex gap-2">
            <button
              onClick={() => setViewMode('all')}
              className={`btn ${viewMode === 'all' ? 'btn-primary' : 'btn-secondary'}`}
            >
              All Documents
            </button>
            <button
              onClick={() => setViewMode('pending')}
              className={`btn ${viewMode === 'pending' ? 'btn-primary' : 'btn-secondary'}`}
            >
              Pending Review
            </button>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left: upload + list */}
        <div className="lg:col-span-1 space-y-6">
          <div className="card">
            <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
              <UploadCloud className="w-5 h-5" /> Upload
            </h3>

            <form onSubmit={onUpload} className="space-y-3">
              <input
                className="input"
                placeholder="Title (optional)"
                value={title}
                onChange={(e) => setTitle(e.target.value)}
              />

              <input
                className="input"
                type="file"
                required
                onChange={(e) => setFile(e.target.files?.[0] || null)}
              />

              <button
                type="submit"
                className="btn btn-primary w-full flex items-center justify-center gap-2"
                disabled={uploading || !file}
              >
                <UploadCloud className="w-4 h-4" />
                {uploading ? 'Uploading…' : 'Upload'}
              </button>

              <p className="text-xs text-white/60">Supported: PDF, TXT, MD, JSON, CSV (max 15MB)</p>
            </form>
          </div>

          <div className="card">
            <h3 className="text-lg font-bold text-white mb-4 flex items-center gap-2">
              <FileText className="w-5 h-5" /> 
              {viewMode === 'pending' ? 'Pending Review' : 'Uploaded'}
            </h3>

            {docs.length === 0 ? (
              <p className="text-sm text-white/70">
                {viewMode === 'pending' ? 'No documents pending review.' : 'No documents uploaded yet.'}
              </p>
            ) : (
              <div className="space-y-2">
                {docs.map((d) => {
                  const active = d.id === selectedId;
                  const status = statusBadge(d.status);
                  const StatusIcon = status.icon;
                  return (
                    <button
                      key={d.id}
                      type="button"
                      onClick={() => {
                        setSelectedId(d.id);
                        setResult(null);
                        setSummary(null);
                      }}
                      className={`w-full text-left rounded-xl border px-4 py-3 transition-all ${
                        active
                          ? 'border-primary-500 bg-slate-800/70'
                          : 'border-slate-800 bg-slate-900/60 hover:bg-slate-800/40'
                      }`}
                    >
                      <div className="flex items-center justify-between gap-3">
                        <div className="min-w-0 flex-1">
                          <div className="font-semibold text-white truncate">
                            {d.title || d.filename}
                          </div>
                          <div className="text-xs text-white/60 truncate">{d.filename}</div>
                          {d.uploaded_by_name && (
                            <div className="text-xs text-white/50 mt-1">By: {d.uploaded_by_name}</div>
                          )}
                        </div>
                        <div className="flex flex-col items-end gap-1">
                          <span className={`px-2 py-0.5 rounded-full text-xs font-medium flex items-center gap-1 ${status.bg}`}>
                            <StatusIcon className="w-3 h-3" />
                            {d.status || 'submitted'}
                          </span>
                          <div className="text-xs text-white/60 whitespace-nowrap">
                            {d.created_at ? new Date(d.created_at).toLocaleDateString() : ''}
                          </div>
                        </div>
                      </div>
                    </button>
                  );
                })}
              </div>
            )}
          </div>
        </div>

        {/* Right: verify & review */}
        <div className="lg:col-span-2 space-y-6">
          {/* Document Details */}
          <div className="card">
            <h3 className="text-lg font-bold text-white mb-2">Document Details</h3>

            <div className="rounded-xl border border-slate-800 bg-slate-900/60 p-4">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <div className="text-xs text-white/60">Title</div>
                  <div className="text-white font-semibold">
                    {selectedDoc ? (selectedDoc.title || selectedDoc.filename) : 'None selected'}
                  </div>
                </div>
                <div>
                  <div className="text-xs text-white/60">Status</div>
                  {selectedDoc && (
                    <span className={`px-2 py-0.5 rounded-full text-xs font-medium ${statusBadge(selectedDoc.status).bg}`}>
                      {selectedDoc.status || 'submitted'}
                    </span>
                  )}
                </div>
                {selectedDoc?.verified_by_name && (
                  <>
                    <div>
                      <div className="text-xs text-white/60">Verified By</div>
                      <div className="text-white">{selectedDoc.verified_by_name}</div>
                    </div>
                    <div>
                      <div className="text-xs text-white/60">Verified At</div>
                      <div className="text-white">
                        {selectedDoc.verified_at ? new Date(selectedDoc.verified_at).toLocaleString() : '-'}
                      </div>
                    </div>
                  </>
                )}
                {selectedDoc?.verification_notes && (
                  <div className="col-span-2">
                    <div className="text-xs text-white/60">Verification Notes</div>
                    <div className="text-white">{selectedDoc.verification_notes}</div>
                  </div>
                )}
              </div>
              {selectedDoc && (
                <div className="text-xs text-white/60 mt-3">SHA256: {selectedDoc.sha256}</div>
              )}
            </div>

            {/* AI Summary for admins/managers */}
            {isAdminOrManager && selectedDoc && (
              <div className="mt-4">
                <button
                  onClick={onGenerateSummary}
                  disabled={summarizing}
                  className="btn btn-secondary flex items-center gap-2"
                >
                  <Sparkles className="w-4 h-4" />
                  {summarizing ? 'Generating Summary…' : 'Generate AI Summary'}
                </button>

                {summary && (
                  <div className="mt-4 rounded-xl border border-slate-800 bg-slate-900/60 p-4 space-y-3">
                    <div>
                      <div className="text-xs text-white/60 mb-1">Summary</div>
                      <div className="text-sm text-white">{summary.summary}</div>
                    </div>
                    {summary.key_points?.length > 0 && (
                      <div>
                        <div className="text-xs text-white/60 mb-1">Key Points</div>
                        <ul className="list-disc list-inside text-sm text-white/80">
                          {summary.key_points.map((p, i) => <li key={i}>{p}</li>)}
                        </ul>
                      </div>
                    )}
                    <div className="flex gap-4 text-sm">
                      <div>
                        <span className="text-white/60">Type:</span>{' '}
                        <span className="text-white">{summary.document_type}</span>
                      </div>
                      <div>
                        <span className="text-white/60">Quality:</span>{' '}
                        <span className="text-white">{summary.quality_assessment}</span>
                      </div>
                      <div>
                        <span className="text-white/60">Recommendation:</span>{' '}
                        <span className={`font-medium ${
                          summary.verification_recommendation === 'approve' ? 'text-emerald-400' :
                          summary.verification_recommendation === 'reject' ? 'text-red-400' : 'text-yellow-400'
                        }`}>{summary.verification_recommendation}</span>
                      </div>
                    </div>
                    {summary.notes && (
                      <div>
                        <div className="text-xs text-white/60 mb-1">Notes</div>
                        <div className="text-sm text-white/80">{summary.notes}</div>
                      </div>
                    )}
                  </div>
                )}
              </div>
            )}

            {/* Status Update for admins/managers */}
            {isAdminOrManager && selectedDoc && selectedDoc.status === 'submitted' && (
              <div className="mt-4 rounded-xl border border-slate-800 bg-slate-900/60 p-4">
                <div className="text-sm font-semibold text-white mb-3">Review Document</div>
                <textarea
                  className="input w-full mb-3"
                  placeholder="Add verification notes (optional)"
                  rows={2}
                  value={statusNotes}
                  onChange={(e) => setStatusNotes(e.target.value)}
                />
                <div className="flex gap-3">
                  <button
                    onClick={() => onUpdateStatus('verified')}
                    disabled={updating}
                    className="btn btn-success flex items-center gap-2"
                  >
                    <CheckCircle className="w-4 h-4" />
                    Verify
                  </button>
                  <button
                    onClick={() => onUpdateStatus('rejected')}
                    disabled={updating}
                    className="btn btn-danger flex items-center gap-2"
                  >
                    <XCircle className="w-4 h-4" />
                    Reject
                  </button>
                </div>
              </div>
            )}
          </div>

          {/* AI Verification */}
          <div className="card">
            <h3 className="text-lg font-bold text-white mb-2">Verify with AI</h3>
            <p className="text-sm text-white/70 mb-4">
              Ask a question about the document content. The AI answers using retrieved chunks.
            </p>

            <form onSubmit={onVerify} className="space-y-3">
              <input
                className="input"
                placeholder="Ask a verification question (e.g., What is the invoice total?)"
                value={question}
                onChange={(e) => setQuestion(e.target.value)}
              />

              <button
                type="submit"
                className="btn btn-success flex items-center gap-2"
                disabled={verifying || !selectedDoc}
              >
                <Search className="w-4 h-4" />
                {verifying ? 'Verifying…' : 'Verify'}
              </button>
            </form>

            {result && (
              <div className="mt-6 space-y-4">
                <div className="flex items-center gap-3 flex-wrap">
                  <span className={`px-3 py-1 rounded-full text-xs font-semibold ${verdictBadge(result.verdict)}`}>
                    {String(result.verdict || 'insufficient').toUpperCase()}
                  </span>
                  <span className="text-sm text-white/70">
                    Confidence: {typeof result.confidence === 'number' ? Math.round(result.confidence * 100) : 0}%
                  </span>
                </div>

                <div className="rounded-xl border border-slate-800 bg-slate-900/60 p-4">
                  <div className="text-xs text-white/60 mb-2">Answer</div>
                  <div className="text-sm text-white whitespace-pre-wrap">{result.answer || ''}</div>
                </div>

                <div className="rounded-xl border border-slate-800 bg-slate-900/60 p-4">
                  <div className="text-xs text-white/60 mb-2">Citations</div>
                  {Array.isArray(result.citations) && result.citations.length > 0 ? (
                    <div className="space-y-2">
                      {result.citations.map((c, idx) => (
                        <div key={`${c.chunk_index}-${idx}`} className="text-sm">
                          <div className="text-white/80 font-semibold">Chunk {c.chunk_index}</div>
                          <div className="text-white/70">{c.snippet}</div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="text-sm text-white/70">No citations returned.</div>
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
