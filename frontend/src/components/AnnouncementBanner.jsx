import { h } from "preact";
import { useState, useEffect } from "preact/hooks";
import api from "../utils/api";
import { getToken } from "../utils/auth";
import AnnouncementToast from "./AnnouncementToast";
import AnnouncementModal from "./AnnouncementModal";

export default function AnnouncementBanner() {
  const [announcements, setAnnouncements] = useState([]);
  const [toastAnnouncement, setToastAnnouncement] = useState(null);
  const [modalAnnouncement, setModalAnnouncement] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (getToken()) {
      loadAnnouncements();
    } else {
      setLoading(false);
    }
  }, []);

  const getDismissedAnnouncements = () => {
    try {
      const dismissed = localStorage.getItem("dismissedAnnouncements");
      return dismissed ? JSON.parse(dismissed) : [];
    } catch {
      return [];
    }
  };

  const markAsDismissed = (announcementId) => {
    try {
      const dismissed = getDismissedAnnouncements();
      if (!dismissed.includes(announcementId)) {
        dismissed.push(announcementId);
        localStorage.setItem(
          "dismissedAnnouncements",
          JSON.stringify(dismissed),
        );
      }
    } catch (err) {
      console.error("Failed to save dismissed announcement:", err);
    }
  };

  const loadAnnouncements = async () => {
    const token = getToken();
    if (!token) {
      setLoading(false);
      return;
    }

    try {
      const response = await api.get("/announcements");
      const allAnnouncements = response.data?.announcements || [];
      const dismissed = getDismissedAnnouncements();

      // Filter out dismissed announcements
      const activeAnnouncements = allAnnouncements.filter(
        (a) => !dismissed.includes(a.id),
      );

      setAnnouncements(activeAnnouncements);

      console.log(
        "[AnnouncementBanner] Loaded announcements:",
        activeAnnouncements.length,
      );

      // Show first announcement as modal (box style)
      if (activeAnnouncements.length > 0) {
        console.log(
          "[AnnouncementBanner] Showing modal:",
          activeAnnouncements[0].title,
        );
        setModalAnnouncement(activeAnnouncements[0]);
      }
    } catch (err) {
      // Silently fail for 401 errors (not authenticated)
      if (err.response?.status === 401) {
        console.log(
          "[AnnouncementBanner] Not authenticated, skipping announcements",
        );
      } else {
        console.error("Failed to load announcements:", err);
      }
    } finally {
      setLoading(false);
    }
  };

  const handleDismiss = async (announcementId) => {
    try {
      await api.post(`/announcements/${announcementId}/dismiss`);
      markAsDismissed(announcementId);
      const remaining = announcements.filter((a) => a.id !== announcementId);
      setAnnouncements(remaining);

      console.log("[AnnouncementBanner] Dismissed:", announcementId);
      console.log(
        "[AnnouncementBanner] Remaining announcements:",
        remaining.length,
      );

      // Clear modal if it's the one being dismissed
      if (modalAnnouncement?.id === announcementId) {
        setModalAnnouncement(null);
        // Show next announcement as toast if available
        if (remaining.length > 0) {
          console.log(
            "[AnnouncementBanner] Showing next announcement as toast:",
            remaining[0].title,
          );
          setTimeout(() => {
            setToastAnnouncement(remaining[0]);
          }, 300);
        } else {
          console.log("[AnnouncementBanner] No more announcements to show");
        }
      }

      // Clear toast if it's the one being dismissed
      if (toastAnnouncement?.id === announcementId) {
        setToastAnnouncement(null);
        // Show next announcement as toast if available
        if (remaining.length > 0) {
          console.log(
            "[AnnouncementBanner] Showing next announcement as toast:",
            remaining[0].title,
          );
          setTimeout(() => {
            setToastAnnouncement(remaining[0]);
          }, 300);
        }
      }
    } catch (err) {
      console.error("Failed to dismiss announcement:", err);
    }
  };

  if (loading) {
    return null;
  }

  return (
    <>
      {modalAnnouncement && (
        <AnnouncementModal
          announcement={modalAnnouncement}
          onClose={handleDismiss}
        />
      )}
      {toastAnnouncement && (
        <AnnouncementToast
          announcement={toastAnnouncement}
          onDismiss={handleDismiss}
        />
      )}
    </>
  );
}
