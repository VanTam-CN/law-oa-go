import React from "react";
import { LinkContainer } from "react-router-bootstrap";
import { useSelector, useDispatch } from "react-redux";
import { useNavigate } from "react-router-dom";
import { RootState } from "../../store";
import { logout } from "../../store/slices/authSlice";
import useTranslation from "../../hooks/useTranslation";
import ThemeToggle from "../ui/ThemeToggle";

const TopNavbar: React.FC = () => {
  const { user } = useSelector((state: RootState) => state.auth);
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { t } = useTranslation();

  const handleLogout = () => {
    dispatch(logout() as any);
    navigate("/login");
  };

  // 获取角色显示文本
  const getRoleText = (role: string) => {
    switch (role) {
      case "admin":
        return t("common.admin");
      case "lawyer":
        return t("common.lawyer");
      case "user":
        return t("common.user");
      default:
        return role;
    }
  };

  return (
    <nav className="bg-card border-b border-border sticky top-0 z-10">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex">
            <div className="flex-shrink-0 flex items-center">
              <div className="bg-primary w-8 h-8 rounded-lg flex items-center justify-center">
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-5 w-5 text-primary-foreground"
                  viewBox="0 0 20 20"
                  fill="currentColor"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 2a1 1 0 011 1v1.323l3.954 1.582 1.599-.8a1 1 0 01.894 1.79l-1.233.616 1.738 5.42a1 1 0 01-.285 1.05A3.989 3.989 0 0115 15a3.989 3.989 0 01-2.667-1.019 1 1 0 01-.285-1.049l1.715-5.349L11 6.477V5h2a1 1 0 110 2H9a1 1 0 01-1-1V3a1 1 0 011-1h1zm-6 8a1 1 0 011 1v1.323l3.954 1.582 1.599-.8a1 1 0 11.894 1.79l-1.233.616 1.738 5.42a1 1 0 01-.285 1.05A3.989 3.989 0 019 21a3.989 3.989 0 01-2.667-1.019 1 1 0 01-.285-1.049l1.715-5.349L5 12.477V11a1 1 0 011-1z"
                    clipRule="evenodd"
                  />
                </svg>
              </div>
              <span className="ml-2 text-xl font-bold text-foreground hidden md:block">
                Law OA System
              </span>
            </div>
            <div className="hidden sm:ml-6 sm:flex sm:space-x-8">
              <LinkContainer to="/dashboard">
                <a className="border-transparent text-foreground hover:border-gray-300 hover:text-muted-foreground inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 mr-2"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"
                    />
                  </svg>
                  {t("navigation.dashboard")}
                </a>
              </LinkContainer>
              <LinkContainer to="/clients">
                <a className="border-transparent text-foreground hover:border-gray-300 hover:text-muted-foreground inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 mr-2"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
                    />
                  </svg>
                  {t("navigation.clients")}
                </a>
              </LinkContainer>
              <LinkContainer to="/cases">
                <a className="border-transparent text-foreground hover:border-gray-300 hover:text-muted-foreground inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 mr-2"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
                    />
                  </svg>
                  {t("navigation.cases")}
                </a>
              </LinkContainer>
              {user?.role === "admin" && (
                <LinkContainer to="/users">
                  <a className="border-transparent text-foreground hover:border-gray-300 hover:text-muted-foreground inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium">
                    <svg
                      xmlns="http://www.w3.org/2000/svg"
                      className="h-5 w-5 mr-2"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"
                      />
                    </svg>
                    {t("navigation.users")}
                  </a>
                </LinkContainer>
              )}
            </div>
          </div>
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <ThemeToggle />
            </div>
            <div className="hidden md:ml-4 md:flex md:flex-shrink-0 md:items-center">
              <div className="ml-3 relative">
                <div className="flex items-center space-x-3">
                  <div className="text-right hidden md:block">
                    <div className="text-sm font-medium text-foreground">
                      {user?.name}
                    </div>
                    <div className="text-xs text-muted-foreground">
                      {getRoleText(user?.role || "")}
                    </div>
                  </div>
                  <div className="flex-shrink-0">
                    <div className="bg-muted w-10 h-10 rounded-full flex items-center justify-center">
                      <span className="text-foreground font-medium">
                        {user?.name.charAt(0).toUpperCase()}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </nav>
  );
};

export default TopNavbar;
