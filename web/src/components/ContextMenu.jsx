import { useEffect, useState, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Trash2, Pause, Play, Link, FolderOpen, X } from 'lucide-react';

const ContextMenuItem = ({ icon: Icon, label, onClick, active = true, danger = false }) => (
    <button
        onClick={onClick}
        className={`w-full flex items-center gap-2 px-3 py-2 text-sm rounded-md transition-colors ${active
            ? danger
                ? 'hover:bg-red-500/20 text-red-400 hover:text-red-300'
                : 'hover:bg-white/10 text-gray-200 hover:text-white'
            : 'text-gray-500 cursor-not-allowed'
            }`}
        disabled={!active}
    >
        <Icon size={16} />
        {label}
    </button>
)

export default function ContextMenu({ onRemove }) {
    const [anchorPoint, setAnchorPoint] = useState({ x: 0, y: 0 });
    const [show, setShow] = useState(false);
    const [selectedTorrent, setSelectedTorrent] = useState(null);

    const handleContextMenu = useCallback((event) => {
        // Check if the target is a torrent item
        const item = event.target.closest('[data-torrent-hash]');
        if (item) {
            event.preventDefault();
            const hash = item.getAttribute('data-torrent-hash');
            const name = item.getAttribute('data-torrent-name');
            setSelectedTorrent({ hash, name });
            setAnchorPoint({ x: event.pageX, y: event.pageY });
            setShow(true);
        } else {
            setShow(false);
        }
    }, []);

    const handleClick = useCallback(() => (show ? setShow(false) : null), [show]);

    const handleRemove = () => {
        if (onRemove && selectedTorrent) {
            onRemove(selectedTorrent.hash);
        }
        setShow(false); // Close menu
    };

    useEffect(() => {
        document.addEventListener('click', handleClick);
        document.addEventListener('contextmenu', handleContextMenu);
        return () => {
            document.removeEventListener('click', handleClick);
            document.removeEventListener('contextmenu', handleContextMenu);
        };
    }, [handleClick, handleContextMenu]);

    if (!show) return null;

    return (
        <AnimatePresence>
            {show && (
                <motion.div
                    initial={{ opacity: 0, scale: 0.9, y: 10 }}
                    animate={{ opacity: 1, scale: 1, y: 0 }}
                    exit={{ opacity: 0, scale: 0.9 }}
                    style={{ top: anchorPoint.y, left: anchorPoint.x }}
                    className="fixed z-50 w-48 bg-[#0a0a0a]/90 backdrop-blur-xl border border-white/10 rounded-xl shadow-2xl p-1.5 flex flex-col gap-0.5 transform origin-top-left"
                >
                    <div className="px-3 py-1.5 text-xs font-bold text-gray-500 border-b border-white/5 mb-1 truncate">
                        {selectedTorrent?.name}
                    </div>

                    <ContextMenuItem icon={Pause} label="Pause" onClick={() => console.log('Pause', selectedTorrent.hash)} />
                    <ContextMenuItem icon={Play} label="Resume" onClick={() => console.log('Resume', selectedTorrent.hash)} active={false} />
                    <ContextMenuItem icon={FolderOpen} label="Open Folder" onClick={() => console.log('Open', selectedTorrent.hash)} />
                    <ContextMenuItem icon={Link} label="Copy Magnet" onClick={() => console.log('Copy', selectedTorrent.hash)} />

                    <div className="h-px bg-white/5 my-1" />

                    <ContextMenuItem icon={Trash2} label="Remove" danger onClick={handleRemove} />
                </motion.div>
            )}
        </AnimatePresence>
    );
}
