import { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import StarfieldBackground from './components/StarfieldBackground';
import TiltCard from './components/TiltCard';
import ContextMenu from './components/ContextMenu';
import { Activity, Download, Upload, Plus, HardDrive, Wifi, Link } from 'lucide-react';

export default function Dashboard() {
    const [stats, setStats] = useState({ download_speed_bps: 0, upload_speed_bps: 0, active_torrents: 0 });
    const [torrents, setTorrents] = useState([]);
    const [showAddModal, setShowAddModal] = useState(false);
    const [magnetLink, setMagnetLink] = useState('');
    const [viewLogsTorrent, setViewLogsTorrent] = useState(null); // For Logs Modal

    // Fetch Logic (same as before)
    useEffect(() => {
        const fetchData = async () => {
            try {
                const [statsRes, torrentsRes] = await Promise.all([
                    fetch('http://localhost:8080/api/stats'),
                    fetch('http://localhost:8080/api/torrents')
                ]);
                if (statsRes.ok) setStats(await statsRes.json());
                if (torrentsRes.ok) setTorrents(await torrentsRes.json());
            } catch (err) {
                console.error(err);
            }
        };
        fetchData();
        const interval = setInterval(fetchData, 2000);
        return () => clearInterval(interval);
    }, []);

    const handleAddTorrent = async () => {
        try {
            await fetch('http://localhost:8080/api/torrents/add', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ magnet_link: magnetLink }),
            });
            setShowAddModal(false);
            setMagnetLink('');
        } catch (err) {
            console.error("Failed to add torrent:", err);
        }
    };

    const handleRemoveTorrent = async (hash) => {
        try {
            await fetch(`http://localhost:8080/api/torrents/remove?hash=${hash}`, {
                method: 'POST',
            });
        } catch (err) {
            console.error("Failed to remove torrent:", err);
        }
    };

    const formatSpeed = (bps) => {
        if (bps === 0) return '0 B/s';
        const k = 1024;
        const sizes = ['B/s', 'KB/s', 'MB/s', 'GB/s'];
        const i = Math.floor(Math.log(bps) / Math.log(k));
        return parseFloat((bps / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    };

    return (
        <>
            <StarfieldBackground speed={stats.download_speed_bps > 0 ? 2 : 0} />
            <ContextMenu onRemove={handleRemoveTorrent} />

            <div className="min-h-screen text-white font-sans overflow-x-hidden selection:bg-violet-500/30">

                {/* Navbar */}
                <nav className="fixed top-0 w-full z-20 border-b border-white/5 bg-black/20 backdrop-blur-xl">
                    <div className="max-w-7xl mx-auto px-6 h-20 flex items-center justify-between">
                        <div className="flex items-center gap-3">
                            <div className="relative group">
                                <div className="absolute -inset-1 bg-gradient-to-r from-blue-600 to-violet-600 rounded-lg blur opacity-40 group-hover:opacity-100 transition duration-200"></div>
                                <div className="relative w-10 h-10 bg-black rounded-lg flex items-center justify-center border border-white/10">
                                    <Wifi className="w-6 h-6 text-violet-400" />
                                </div>
                            </div>
                            <span className="font-bold text-2xl tracking-tighter bg-clip-text text-transparent bg-gradient-to-r from-white to-gray-400">
                                GoTorrent
                            </span>
                        </div>
                        <motion.button
                            whileHover={{ scale: 1.05 }}
                            whileTap={{ scale: 0.95 }}
                            onClick={() => setShowAddModal(true)}
                            className="flex items-center gap-2 bg-white/10 hover:bg-white/20 border border-white/10 hover:border-violet-500/50 text-white px-5 py-2.5 rounded-xl backdrop-blur-md transition-all shadow-lg"
                        >
                            <Plus size={18} /> <span className="font-medium">Add Torrent</span>
                        </motion.button>
                    </div>
                </nav>

                <main className="pt-32 pb-12 px-6 max-w-7xl mx-auto">
                    {/* Stats Grid */}
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-8 mb-16 perspective-1000">
                        <TiltCard className="p-6">
                            <div className="flex items-start justify-between">
                                <div>
                                    <p className="text-gray-400 text-sm font-medium mb-1">Download Speed</p>
                                    <h3 className="text-4xl font-bold tracking-tight text-white">{formatSpeed(stats.download_speed_bps)}</h3>
                                </div>
                                <div className="p-3 bg-green-500/20 rounded-xl text-green-400 shadow-[0_0_15px_rgba(74,222,128,0.2)]">
                                    <Download size={24} />
                                </div>
                            </div>
                            <div className="mt-4 h-1 w-full bg-white/5 rounded-full overflow-hidden">
                                <motion.div
                                    className="h-full bg-green-500 shadow-[0_0_10px_#22c55e]"
                                    initial={{ width: 0 }}
                                    animate={{ width: `${Math.min((stats.download_speed_bps / 10000000) * 100, 100)}%` }} // Dummy scale relative to 10MB/s
                                    transition={{ type: "spring", stiffness: 50 }}
                                />
                            </div>
                        </TiltCard>

                        <TiltCard className="p-6">
                            <div className="flex items-start justify-between">
                                <div>
                                    <p className="text-gray-400 text-sm font-medium mb-1">Upload Speed</p>
                                    <h3 className="text-4xl font-bold tracking-tight text-white">{formatSpeed(stats.upload_speed_bps)}</h3>
                                </div>
                                <div className="p-3 bg-blue-500/20 rounded-xl text-blue-400 shadow-[0_0_15px_rgba(96,165,250,0.2)]">
                                    <Upload size={24} />
                                </div>
                            </div>
                            <div className="mt-4 h-1 w-full bg-white/5 rounded-full overflow-hidden">
                                <motion.div
                                    className="h-full bg-blue-500 shadow-[0_0_10px_#3b82f6]"
                                    initial={{ width: 0 }}
                                    animate={{ width: `${Math.min((stats.upload_speed_bps / 5000000) * 100, 100)}%` }}
                                    transition={{ type: "spring", stiffness: 50 }}
                                />
                            </div>
                        </TiltCard>

                        <TiltCard className="p-6">
                            <div className="flex items-start justify-between">
                                <div>
                                    <p className="text-gray-400 text-sm font-medium mb-1">Active Tasks</p>
                                    <h3 className="text-4xl font-bold tracking-tight text-white">{stats.active_torrents}</h3>
                                </div>
                                <div className="p-3 bg-violet-500/20 rounded-xl text-violet-400 shadow-[0_0_15px_rgba(139,92,246,0.2)]">
                                    <Activity size={24} />
                                </div>
                            </div>
                            <div className="mt-4 flex items-center gap-2 text-xs text-gray-500">
                                <span className="w-2 h-2 bg-green-500 rounded-full animate-pulse"></span>
                                System Operational
                            </div>
                        </TiltCard>
                    </div>

                    {/* List Header */}
                    <div className="flex items-center gap-4 mb-6">
                        <h2 className="text-2xl font-bold text-white">Active Downloads</h2>
                        <div className="h-px flex-1 bg-gradient-to-r from-white/10 to-transparent"></div>
                    </div>

                    {/* List */}
                    <div className="space-y-4">
                        <AnimatePresence>
                            {torrents.map((t) => (
                                <motion.div
                                    key={t.hash}
                                    initial={{ opacity: 0, y: 20 }}
                                    animate={{ opacity: 1, y: 0 }}
                                    exit={{ opacity: 0, x: -20 }}
                                    layout
                                    data-torrent-hash={t.hash}
                                    data-torrent-name={t.name}
                                    onClick={(e) => {
                                        // Prevent click if context menu was triggered or text selected
                                        if (e.defaultPrevented) return;
                                        setViewLogsTorrent(t);
                                    }}
                                >
                                    <TiltCard className="p-4 group cursor-pointer hover:border-violet-500/30 transition-colors">
                                        <div className="flex items-center gap-6">
                                            <div className="w-12 h-12 rounded-xl bg-gradient-to-br from-gray-800 to-black border border-white/10 flex items-center justify-center shadow-inner">
                                                <HardDrive className="text-gray-400 group-hover:text-white transition-colors" />
                                            </div>
                                            <div className="flex-1 min-w-0">
                                                <div className="flex justify-between items-center mb-2">
                                                    <h4 className="font-semibold text-lg truncate pr-4">{t.name}</h4>
                                                    <span className="font-mono text-sm text-gray-400">{t.progress.toFixed(1)}%</span>
                                                </div>

                                                <div className="h-2 w-full bg-black/50 rounded-full overflow-hidden mb-3 border border-white/5">
                                                    <motion.div
                                                        className="h-full bg-gradient-to-r from-blue-600 via-violet-600 to-fuchsia-600"
                                                        initial={{ width: 0 }}
                                                        animate={{ width: `${t.progress}%` }}
                                                    />
                                                </div>

                                                <div className="flex items-center gap-6 text-xs font-medium text-gray-500">
                                                    <span className="flex items-center gap-1.5">
                                                        <span className={`w-1.5 h-1.5 rounded-full ${t.state === 'Downloading' ? 'bg-green-500 shadow-[0_0_8px_#22c55e]' : 'bg-yellow-500'}`}></span>
                                                        {t.state}
                                                    </span>
                                                    <span className="group-hover:text-blue-400 transition-colors">{formatSpeed(t.speed_in)} ↓</span>
                                                    <span className="group-hover:text-violet-400 transition-colors">{t.peers} peers</span>
                                                </div>
                                            </div>
                                        </div>
                                    </TiltCard>
                                </motion.div>
                            ))}
                        </AnimatePresence>
                    </div>
                </main>

                {/* Modal */}
                <AnimatePresence>
                    {showAddModal && (
                        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
                            <motion.div
                                initial={{ opacity: 0 }}
                                animate={{ opacity: 1 }}
                                exit={{ opacity: 0 }}
                                onClick={() => setShowAddModal(false)}
                                className="absolute inset-0 bg-black/80 backdrop-blur-sm"
                            />
                            <motion.div
                                initial={{ scale: 0.9, opacity: 0, y: 20 }}
                                animate={{ scale: 1, opacity: 1, y: 0 }}
                                exit={{ scale: 0.9, opacity: 0, y: 20 }}
                                className="relative w-full max-w-lg bg-[#0a0a0a] border border-white/10 rounded-2xl p-8 shadow-2xl overflow-hidden"
                            >
                                {/* Modal Glow */}
                                <div className="absolute top-0 inset-x-0 h-px bg-gradient-to-r from-transparent via-violet-500 to-transparent opacity-50"></div>

                                <h3 className="text-2xl font-bold mb-6 text-white text-center">Add New Torrent</h3>

                                <div className="space-y-6">
                                    <div className="relative">
                                        <input
                                            type="text"
                                            className="w-full bg-white/5 border border-white/10 rounded-xl px-5 py-4 text-white focus:outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-500 transition-all placeholder-gray-600"
                                            placeholder="Paste magnet link here..."
                                            value={magnetLink}
                                            onChange={(e) => setMagnetLink(e.target.value)}
                                            autoFocus
                                        />
                                        <div className="absolute right-4 top-1/2 -translate-y-1/2">
                                            <Link className="text-gray-500" size={18} />
                                        </div>
                                    </div>

                                    <div className="flex gap-4">
                                        <button
                                            onClick={() => setShowAddModal(false)}
                                            className="flex-1 py-3 rounded-xl border border-white/10 text-gray-400 hover:text-white hover:bg-white/5 transition-all font-medium"
                                        >
                                            Cancel
                                        </button>
                                        <button
                                            onClick={handleAddTorrent}
                                            disabled={!magnetLink}
                                            className="flex-1 py-3 rounded-xl bg-violet-600 hover:bg-violet-500 text-white shadow-lg shadow-violet-600/20 disabled:opacity-50 disabled:shadow-none transition-all font-bold"
                                        >
                                            Download Magnet
                                        </button>
                                    </div>

                                    <div className="relative">
                                        <div className="absolute inset-0 flex items-center">
                                            <div className="w-full border-t border-white/10"></div>
                                        </div>
                                        <div className="relative flex justify-center text-sm">
                                            <span className="px-2 bg-[#0a0a0a] text-gray-500">Or from file</span>
                                        </div>
                                    </div>

                                    <div className="flex justify-center">
                                        <input
                                            type="file"
                                            id="torrent-file-upload"
                                            className="hidden"
                                            accept=".torrent"
                                            onChange={async (e) => {
                                                const file = e.target.files[0];
                                                if (!file) return;

                                                const formData = new FormData();
                                                formData.append('torrent_file', file);

                                                try {
                                                    const res = await fetch('http://localhost:8080/api/torrents/upload', {
                                                        method: 'POST',
                                                        body: formData,
                                                    });
                                                    if (res.ok) {
                                                        setShowAddModal(false);
                                                        setMagnetLink(''); // Clear magnet
                                                    } else {
                                                        console.error("Upload failed");
                                                    }
                                                } catch (err) {
                                                    console.error("Upload error", err);
                                                }
                                            }}
                                        />
                                        <label
                                            htmlFor="torrent-file-upload"
                                            className="cursor-pointer flex items-center gap-2 px-6 py-3 rounded-xl bg-white/5 hover:bg-white/10 border border-white/10 transition-all group"
                                        >
                                            <HardDrive className="text-gray-400 group-hover:text-white" size={18} />
                                            <span className="text-gray-300 group-hover:text-white font-medium">Select .torrent File</span>
                                        </label>
                                    </div>
                                </div>
                            </motion.div>
                        </div>
                    )}
                </AnimatePresence>

                {/* Logs Modal */}
                <AnimatePresence>
                    {viewLogsTorrent && (
                        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
                            <motion.div
                                initial={{ opacity: 0 }}
                                animate={{ opacity: 1 }}
                                exit={{ opacity: 0 }}
                                onClick={() => setViewLogsTorrent(null)}
                                className="absolute inset-0 bg-black/90 backdrop-blur-sm"
                            />
                            <motion.div
                                initial={{ scale: 0.95, opacity: 0, y: 20 }}
                                animate={{ scale: 1, opacity: 1, y: 0 }}
                                exit={{ scale: 0.95, opacity: 0, y: 20 }}
                                className="relative w-full max-w-2xl bg-[#0d0d0d] border border-white/10 rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[80vh]"
                            >
                                {/* Header */}
                                <div className="p-6 border-b border-white/5 flex items-center justify-between bg-white/5">
                                    <div>
                                        <h3 className="text-xl font-bold text-white truncate max-w-md">{viewLogsTorrent.name}</h3>
                                        <div className="flex items-center gap-2 text-sm text-gray-400 mt-1">
                                            <span className={`w-2 h-2 rounded-full ${viewLogsTorrent.state.includes('Stalled') ? 'bg-red-500' : 'bg-green-500 animate-pulse'}`}></span>
                                            {viewLogsTorrent.state}
                                        </div>
                                    </div>
                                    <button
                                        onClick={() => setViewLogsTorrent(null)}
                                        className="p-2 hover:bg-white/10 rounded-lg transition-colors"
                                    >
                                        <span className="text-gray-400 hover:text-white">Esc</span>
                                    </button>
                                </div>

                                {/* Logs Console */}
                                <div className="flex-1 overflow-y-auto p-6 font-mono text-sm bg-black/50 space-y-2 scrollbar-thin scrollbar-thumb-white/10 scrollbar-track-transparent">
                                    {viewLogsTorrent.logs && viewLogsTorrent.logs.length > 0 ? (
                                        viewLogsTorrent.logs.map((log, i) => (
                                            <div key={i} className="flex gap-3 text-gray-300 border-l-2 border-white/5 pl-3 hover:border-violet-500/50 hover:bg-white/5 transition-colors p-1 rounded-r">
                                                <span className="text-violet-500 shrink-0 opacity-50">[{i + 1}]</span>
                                                <span className="break-all">{log}</span>
                                            </div>
                                        ))
                                    ) : (
                                        <div className="text-gray-600 italic text-center py-10">No logs available (Client initialized...)</div>
                                    )}
                                    {/* Auto-scroll dummy */}
                                    <div ref={(el) => el?.scrollIntoView({ behavior: 'smooth' })} />
                                </div>
                            </motion.div>
                        </div>
                    )}
                </AnimatePresence>
            </div>
        </>
    );
}
